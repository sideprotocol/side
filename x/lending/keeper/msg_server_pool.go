package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// CreatePool implements types.MsgServer.
func (m msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	if m.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if m.HasPool(ctx, msg.Id) {
		return nil, types.ErrPoolAlreadyExists
	}

	yTokenDenom := types.YTokenDenom(msg.Id)
	if m.bankKeeper.HasSupply(ctx, yTokenDenom) {
		return nil, errorsmod.Wrapf(types.ErrInvalidPoolId, "denom %s already exists", yTokenDenom)
	}

	pool := &types.LendingPool{
		Id:           msg.Id,
		Supply:       sdk.NewCoin(msg.Config.LendingAsset.Denom, sdkmath.ZeroInt()),
		TotalYTokens: sdk.NewCoin(yTokenDenom, sdkmath.ZeroInt()),
		Tranches:     types.NewTranches(msg.Config.Tranches),
		Config:       msg.Config,
		Status:       types.PoolStatus_INACTIVE,
	}

	m.SetPool(ctx, pool)

	m.EmitEvent(ctx, msg.Authority)

	return &types.MsgCreatePoolResponse{}, nil
}

// AddLiquidity implements types.MsgServer.
func (m msgServer) AddLiquidity(goCtx context.Context, msg *types.MsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasPool(ctx, msg.PoolId) {
		return nil, types.ErrPoolDoesNotExist
	}

	pool := m.GetPool(ctx, msg.PoolId)
	if pool.Status == types.PoolStatus_PAUSED {
		return nil, types.ErrPoolPaused
	}

	if msg.Amount.Denom != pool.Config.LendingAsset.Denom {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "mismatched denom")
	}

	if err := types.CheckSupplyCap(pool, msg.Amount.Amount); err != nil {
		return nil, types.ErrSupplyCapExceeded
	}

	var yTokenAmount sdkmath.Int

	if pool.Supply.IsZero() {
		// activate pool on first deposit
		pool.Status = types.PoolStatus_ACTIVE
		yTokenAmount = msg.Amount.Amount
	} else {
		yTokenAmount = m.GetYTokenAmount(ctx, pool, msg.Amount.Amount)
	}

	pool.Supply = pool.Supply.Add(msg.Amount)
	pool.AvailableAmount = pool.AvailableAmount.Add(msg.Amount.Amount)
	pool.TotalYTokens = pool.TotalYTokens.AddAmount(yTokenAmount)

	m.SetPool(ctx, pool)

	yTokens := sdk.NewCoin(types.YTokenDenom(pool.Id), yTokenAmount)

	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(msg.Lender), types.ModuleName, sdk.NewCoins(msg.Amount)); err != nil {
		return nil, err
	}

	if err := m.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(yTokens)); err != nil {
		return nil, err
	}

	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(msg.Lender), sdk.NewCoins(yTokens)); err != nil {
		return nil, err
	}

	// Emit Events
	m.EmitEvent(ctx, msg.Lender,
		sdk.NewAttribute("deposit", msg.Amount.String()),
		sdk.NewAttribute("shares", yTokens.String()),
	)

	return &types.MsgAddLiquidityResponse{}, nil
}

// RemoveLiquidity implements types.MsgServer.
func (m msgServer) RemoveLiquidity(goCtx context.Context, msg *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	poolId := types.PoolIdFromYTokenDenom(msg.YTokens.Denom)
	if !m.HasPool(ctx, poolId) {
		return nil, types.ErrPoolDoesNotExist
	}

	pool := m.GetPool(ctx, poolId)
	if pool.Status != types.PoolStatus_ACTIVE {
		return nil, types.ErrPoolNotActive
	}

	var withdrawAmount = m.GetUnderlyingAssetAmount(ctx, pool, msg.YTokens.Amount)
	if withdrawAmount.GT(pool.AvailableAmount) {
		return nil, types.ErrInsufficientLiquidity
	}

	pool.Supply = pool.Supply.SubAmount(withdrawAmount)
	pool.AvailableAmount = pool.AvailableAmount.Sub(withdrawAmount)
	pool.TotalYTokens = pool.TotalYTokens.Sub(msg.YTokens)

	m.SetPool(ctx, pool)

	withdrawAsset := sdk.NewCoin(pool.Supply.Denom, withdrawAmount)

	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(msg.Lender), types.ModuleName, sdk.NewCoins(msg.YTokens)); err != nil {
		return nil, err
	}

	if err := m.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(msg.YTokens)); err != nil {
		return nil, err
	}

	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(msg.Lender), sdk.NewCoins(withdrawAsset)); err != nil {
		return nil, err
	}

	// Emit Events
	m.EmitEvent(ctx, msg.Lender,
		sdk.NewAttribute("burn", msg.YTokens.String()),
		sdk.NewAttribute("withdraw", withdrawAsset.String()),
	)

	return &types.MsgRemoveLiquidityResponse{}, nil
}

// UpdatePoolConfig implements types.MsgServer.
func (m msgServer) UpdatePoolConfig(goCtx context.Context, msg *types.MsgUpdatePoolConfig) (*types.MsgUpdatePoolConfigResponse, error) {
	if m.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasPool(ctx, msg.PoolId) {
		return nil, types.ErrPoolDoesNotExist
	}

	pool := m.GetPool(ctx, msg.PoolId)
	if err := types.ValidatePoolConfigUpdate(pool.Config, msg.Config); err != nil {
		return nil, err
	}

	m.UpdatePoolStatus(ctx, pool, &msg.Config)
	m.OnPoolTranchesConfigChanged(ctx, pool, msg.Config.Tranches)

	pool.Config = msg.Config
	m.SetPool(ctx, pool)

	return &types.MsgUpdatePoolConfigResponse{}, nil
}
