package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/sideprotocol/side/x/farming/types"
)

type msgServer struct {
	Keeper
}

// Stake implements types.MsgServer.
func (m msgServer) Stake(goCtx context.Context, msg *types.MsgStake) (*types.MsgStakeResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.FarmingEnabled(ctx) {
		return nil, types.ErrFarmingNotEnabled
	}

	if !m.IsEligibleAsset(ctx, msg.Amount.Denom) {
		return nil, errorsmod.Wrapf(types.ErrAssetNotEligible, "asset %s not eligible", msg.Amount.Denom)
	}

	if msg.LockDuration < m.EpochDuration(ctx) {
		return nil, errorsmod.Wrapf(types.ErrInvalidLockDuration, "lock duration cannot be less than %s day(s)", types.GetLockDurationInDays(m.EpochDuration(ctx)))
	}

	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(msg.Staker), types.ModuleName, sdk.NewCoins(msg.Amount)); err != nil {
		return nil, err
	}

	lockMultiplier := types.GetLockMultiplier(msg.LockDuration)

	staking := &types.Staking{
		Id:              m.IncrementStakingId(ctx),
		Address:         msg.Staker,
		Amount:          msg.Amount,
		LockDuration:    msg.LockDuration,
		LockMultiplier:  lockMultiplier,
		EffectiveAmount: types.GetEffectiveAmount(msg.Amount, lockMultiplier),
		PendingRewards:  sdk.NewCoin(m.RewardsPerEpoch(ctx).Denom, sdkmath.ZeroInt()),
		StartTime:       ctx.BlockTime(),
		Status:          types.StakingStatus_STAKING_STATUS_STAKED,
	}

	// set staking
	m.SetStaking(ctx, staking)
	m.SetStakingByAddress(ctx, msg.Staker, staking)

	// update total staking
	m.IncreaseTotalStaking(ctx, staking)

	return &types.MsgStakeResponse{}, nil
}

// Unstake implements types.MsgServer.
func (m msgServer) Unstake(goCtx context.Context, msg *types.MsgUnstake) (*types.MsgUnstakeResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasStaking(ctx, msg.Id) {
		return nil, errorsmod.Wrapf(types.ErrStakingDoesNotExist, "id: %d", msg.Id)
	}

	staking := m.GetStaking(ctx, msg.Id)
	if staking.Address != msg.Staker {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "mismatched staker address")
	}

	if staking.Status == types.StakingStatus_STAKING_STATUS_UNSTAKED {
		return nil, errorsmod.Wrapf(types.ErrInvalidStakingStatus, "already unstaked: %d", msg.Id)
	}

	if ctx.BlockTime().Before(staking.StartTime.Add(staking.LockDuration)) {
		return nil, errorsmod.Wrapf(types.ErrLockDurationNotEnded, "lock duration end time: %s", staking.StartTime.Add(staking.LockDuration))
	}

	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(msg.Staker), sdk.NewCoins(staking.Amount)); err != nil {
		return nil, err
	}

	// claim pending rewards if any
	if staking.PendingRewards.IsPositive() {
		if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(msg.Staker), sdk.NewCoins(staking.PendingRewards)); err != nil {
			return nil, err
		}

		// reset pending rewards
		staking.PendingRewards = sdk.NewCoin(staking.PendingRewards.Denom, sdkmath.ZeroInt())
	}

	// update status
	staking.Status = types.StakingStatus_STAKING_STATUS_UNSTAKED
	m.SetStaking(ctx, staking)

	// update total staking
	m.DecreaseTotalStaking(ctx, staking)

	return &types.MsgUnstakeResponse{}, nil
}

// Claim implements types.MsgServer.
func (m msgServer) Claim(goCtx context.Context, msg *types.MsgClaim) (*types.MsgClaimResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasStaking(ctx, msg.Id) {
		return nil, errorsmod.Wrapf(types.ErrStakingDoesNotExist, "id: %d", msg.Id)
	}

	staking := m.GetStaking(ctx, msg.Id)
	if staking.Address != msg.Staker {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "mismatched staker address")
	}

	if staking.Status == types.StakingStatus_STAKING_STATUS_UNSTAKED {
		return nil, errorsmod.Wrapf(types.ErrInvalidStakingStatus, "already unstaked: %d", msg.Id)
	}

	if staking.PendingRewards.IsZero() {
		return nil, types.ErrNoPendingRewards
	}

	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(msg.Staker), sdk.NewCoins(staking.PendingRewards)); err != nil {
		return nil, err
	}

	// reset pending rewards
	staking.PendingRewards = sdk.NewCoin(staking.PendingRewards.Denom, sdkmath.ZeroInt())
	m.SetStaking(ctx, staking)

	return &types.MsgClaimResponse{}, nil
}

// UpdateParams updates the module params.
func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if m.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// get the current params
	params := m.GetParams(ctx)

	// update params
	m.SetParams(ctx, msg.Params)

	// handle params change
	m.OnParamsChanged(ctx, params, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
