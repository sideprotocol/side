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

// CreatePhase implements types.MsgServer.
func (m msgServer) CreatePhase(goCtx context.Context, msg *types.MsgCreatePhase) (*types.MsgCreatePhaseResponse, error) {
	if m.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !ctx.BlockTime().Before(msg.StartTime.Add(msg.Duration)) {
		return nil, errorsmod.Wrap(types.ErrInvalidPhaseParams, "phase end time has reached")
	}

	phase := &types.Phase{
		Id:                   m.IncrementPhaseId(ctx),
		StartTime:            msg.StartTime,
		Duration:             msg.Duration,
		DistributionInterval: msg.DistributionInterval,
		RewardsPerInterval:   msg.RewardsPerInterval,
		LockDurations:        msg.LockDurations,
		AllowedAssets:        msg.AllowedAssets,
		Status:               types.PhaseStatus_PHASE_STATUS_PENDING,
	}

	if !ctx.BlockTime().Before(phase.StartTime) {
		phase.Status = types.PhaseStatus_PHASE_STATUS_STARTED
	}

	// set phase
	m.SetPhase(ctx, phase)

	return &types.MsgCreatePhaseResponse{}, nil
}

// Stake implements types.MsgServer.
func (m msgServer) Stake(goCtx context.Context, msg *types.MsgStake) (*types.MsgStakeResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasPhase(ctx, msg.PhaseId) {
		return nil, errorsmod.Wrapf(types.ErrPhaseDoesNotExist, "id: %d", msg.PhaseId)
	}

	phase := m.GetPhase(ctx, msg.PhaseId)

	if ctx.BlockTime().Before(phase.StartTime) {
		return nil, errorsmod.Wrapf(types.ErrPhaseNotStarted, "start time: %s", phase.StartTime)
	}

	if !ctx.BlockTime().Before(phase.StartTime.Add(phase.Duration)) {
		return nil, errorsmod.Wrapf(types.ErrPhaseEnded, "end time: %s", phase.StartTime.Add(phase.Duration))
	}

	if !types.IsAllowedAsset(phase, msg.Amount.Denom) {
		return nil, errorsmod.Wrapf(types.ErrAssetNotAllowed, "asset %s not allowed", msg.Amount.Denom)
	}

	if !types.LockDurationExists(phase, msg.LockDuration) {
		return nil, types.ErrInvalidLockDuration
	}

	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(msg.Staker), types.ModuleName, sdk.NewCoins(msg.Amount)); err != nil {
		return nil, err
	}

	phaseRemainingDuration := phase.StartTime.Add(phase.Duration).Sub(ctx.BlockTime())
	lockDuration := min(msg.LockDuration, phaseRemainingDuration)

	lockMultiplier := types.GetLockMultiplier(lockDuration)

	staking := &types.Staking{
		Id:              m.IncrementStakingId(ctx),
		PhaseId:         msg.PhaseId,
		Address:         msg.Staker,
		Amount:          msg.Amount,
		LockDuration:    lockDuration,
		LockMultiplier:  lockMultiplier,
		EffectiveAmount: types.GetEffectiveAmount(msg.Amount, lockMultiplier),
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

	// update status
	staking.Status = types.StakingStatus_STAKING_STATUS_UNSTAKED
	m.SetStaking(ctx, staking)

	// update total staking
	m.DecreaseTotalStaking(ctx, staking)

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
