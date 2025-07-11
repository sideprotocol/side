package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
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

	if ctx.BlockTime().Before(m.PhaseStartTime(ctx)) {
		return nil, types.ErrPhaseNotStarted
	}

	if !ctx.BlockTime().Before(m.PhaseStartTime(ctx).Add(m.PhaseDuration(ctx))) {
		return nil, types.ErrPhaseEnded
	}

	if !m.LockDurationExists(ctx, msg.LockDuration) {
		return nil, types.ErrInvalidLockDuration
	}

	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(msg.Staker), types.ModuleName, sdk.NewCoins(msg.Amount)); err != nil {
		return nil, err
	}

	phaseRemainingDuration := m.PhaseStartTime(ctx).Add(m.PhaseDuration(ctx)).Sub(ctx.BlockTime())
	lockDuration := min(msg.LockDuration, phaseRemainingDuration)

	lockMultiplier := types.GetLockMultiplier(lockDuration)

	staking := &types.Staking{
		Id:              m.IncrementStakingId(ctx),
		Amount:          msg.Amount,
		LockDuration:    lockDuration,
		LockMultiplier:  lockMultiplier,
		EffectiveAmount: types.GetEffectiveAmount(msg.Amount, lockMultiplier),
		StartTime:       ctx.BlockTime(),
		EndTime:         ctx.BlockTime().Add(lockDuration),
		Status:          types.StakingStatus_STAKING_STATUS_STAKED,
	}

	// set staking
	m.SetStaking(ctx, staking)
	m.SetStakingByAddress(ctx, msg.Staker, staking)

	// update total stakings
	m.IncreaseTotalStakings(ctx, staking)

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
	if staking.Status != types.StakingStatus_STAKING_STATUS_STAKED {
		return nil, errorsmod.Wrapf(types.ErrInvalidStakingStatus, "already unstaked: %d", msg.Id)
	}

	if ctx.BlockTime().Before(staking.EndTime) {
		return nil, errorsmod.Wrapf(types.ErrLockDurationNotEnded, "lock duration end time: %s", staking.EndTime)
	}

	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(msg.Staker), sdk.NewCoins(staking.Amount)); err != nil {
		return nil, err
	}

	// update status
	staking.Status = types.StakingStatus_STAKING_STATUS_UNSTAKED
	m.SetStaking(ctx, staking)

	// update total stakings
	m.DecreaseTotalStakings(ctx, staking)

	return &types.MsgUnstakeResponse{}, nil
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

	m.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
