package keeper

import (
	"context"

	"github.com/sideprotocol/side/x/lending/types"
)

type msgServer struct {
	Keeper
}

// AddLiquidity implements types.MsgServer.
func (m msgServer) AddLiquidity(context.Context, *types.MsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	panic("unimplemented")
}

// RemoveLiquidity implements types.MsgServer.
func (m msgServer) RemoveLiquidity(context.Context, *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	panic("unimplemented")
}

// CreateLoan implements types.MsgServer.
func (m msgServer) Apply(context.Context, *types.MsgApply) (*types.MsgApplyResponse, error) {
	panic("unimplemented")
}

// Repay implements types.MsgServer.
func (m msgServer) Repay(context.Context, *types.MsgRepay) (*types.MsgRepayResponse, error) {
	panic("unimplemented")
}

// RequestVaultAddress implements types.MsgServer.
func (m msgServer) Redeem(context.Context, *types.MsgRedeem) (*types.MsgRedeemResponse, error) {
	panic("unimplemented")
}

// SubmitFundingTx implements types.MsgServer.
func (m msgServer) Fund(context.Context, *types.MsgFund) (*types.MsgFundResponse, error) {
	panic("unimplemented")
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
