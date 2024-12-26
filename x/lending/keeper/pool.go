package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/lending/types"
)

// CreatePool implements types.MsgServer.
func (m msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

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
