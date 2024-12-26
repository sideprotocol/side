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
		TotalShares:    0,
		BorrowedAmount: 0,
		Status:         types.PoolStatus_INACTIVE,
	}

	m.SetPool(ctx, pool)

	m.EmitEvent(ctx, msg.Creator)

	return &types.MsgCreatePoolResponse{}, nil
}

// AddLiquidity implements types.MsgServer.
func (m msgServer) AddLiquidity(goCtx context.Context, msg *types.MsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Emit Events
	m.EmitEvent(ctx, msg.Lender) // sdk.NewAttribute("blockhash", msg.Blockhash),
	// sdk.NewAttribute("txid", txHash.String()),
	// sdk.NewAttribute("recipient", recipient.EncodeAddress()),
	return &types.MsgAddLiquidityResponse{}, nil

}

// RemoveLiquidity implements types.MsgServer.
func (m msgServer) RemoveLiquidity(ctx context.Context, msg *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	panic("unimplemented")
}
