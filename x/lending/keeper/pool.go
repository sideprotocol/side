package keeper

import (
	"context"
	"slices"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/lending/types"
)

// CreatePool implements types.MsgServer.
func (m msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	params := m.GetParams(ctx)

	if !slices.Contains(params.PoolCreators, msg.Creator) {
		return nil, types.ErrNotAuthorized
	}

	if m.HasPool(ctx, msg.PoolId) {
		return nil, types.ErrDuplicatedPoolId
	}

	if m.bankKeeper.HasSupply(ctx, msg.PoolId) {
		return nil, types.ErrDuplicatedPoolId
	}

	supply := sdk.NewCoin(msg.LendingAsset, math.NewInt(0))
	pool := types.LendingPool{
		Id:             msg.PoolId,
		Supply:         &supply,
		TotalShares:    math.NewInt(0),
		BorrowedAmount: math.NewInt(0),
		Status:         types.PoolStatus_INACTIVE,
	}

	m.SetPool(ctx, pool)

	m.EmitEvent(ctx, msg.Creator)

	return &types.MsgCreatePoolResponse{}, nil
}

// AddLiquidity implements types.MsgServer.
func (m msgServer) AddLiquidity(goCtx context.Context, msg *types.MsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lender, err2 := sdk.AccAddressFromBech32(msg.Lender)
	if err2 != nil {
		return nil, err2
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if !m.HasPool(ctx, msg.PoolId) {
		return nil, types.ErrPoolDoesNotExist
	}

	pool := m.GetPool(ctx, msg.PoolId)

	if msg.Amount.Denom != pool.Supply.Denom {
		return nil, types.ErrInvalidAmount
	}

	var outAmount math.Int
	if pool.Supply.Amount.Equal(math.NewInt(0)) {
		// active pool on first deposit
		pool.Status = types.PoolStatus_ACTIVE
		outAmount = msg.Amount.Amount
	} else {
		outAmount = msg.Amount.Amount.Mul(pool.TotalShares).Quo(pool.Supply.Amount)
	}
	if pool.Status != types.PoolStatus_ACTIVE {
		return nil, types.ErrInactivePool
	}

	pool.TotalShares = pool.TotalShares.Add(outAmount)
	pool.Supply.Add(*msg.Amount)

	received_shares := sdk.NewCoin(pool.Id, outAmount)

	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, lender, types.ModuleName, sdk.NewCoins(*msg.Amount)); err != nil {
		return nil, err
	}

	if err := m.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(received_shares)); err != nil {
		return nil, err
	}

	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lender, sdk.NewCoins(received_shares)); err != nil {
		return nil, err
	}

	m.SetPool(ctx, pool)

	// Emit Events
	m.EmitEvent(ctx, msg.Lender,
		sdk.NewAttribute("deposit", msg.Amount.String()),
		sdk.NewAttribute("received_share", received_shares.String()),
	)
	return &types.MsgAddLiquidityResponse{
		Shares: &received_shares,
	}, nil

}

// RemoveLiquidity implements types.MsgServer.
func (m msgServer) RemoveLiquidity(goCtx context.Context, msg *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lender, err2 := sdk.AccAddressFromBech32(msg.Lender)
	if err2 != nil {
		return nil, err2
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if !m.HasPool(ctx, msg.Shares.Denom) {
		return nil, types.ErrPoolDoesNotExist
	}

	pool := m.GetPool(ctx, msg.Shares.Denom)

	var outAmount = msg.Shares.Amount.Quo(pool.TotalShares).Mul(pool.Supply.Amount)
	pool.TotalShares = pool.TotalShares.Sub(msg.Shares.Amount)

	withdraw := sdk.NewCoin(pool.Supply.Denom, outAmount)
	pool.Supply.Sub(withdraw)

	m.SetPool(ctx, pool)

	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, lender, types.ModuleName, sdk.NewCoins(*msg.Shares)); err != nil {
		return nil, err
	}

	if err := m.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(*msg.Shares)); err != nil {
		return nil, err
	}

	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lender, sdk.NewCoins(withdraw)); err != nil {
		return nil, err
	}

	// Emit Events
	m.EmitEvent(ctx, msg.Lender,
		sdk.NewAttribute("burn", msg.Shares.String()),
		sdk.NewAttribute("withdraw", withdraw.String()),
	)
	return &types.MsgRemoveLiquidityResponse{
		Amount: &withdraw,
	}, nil
}
