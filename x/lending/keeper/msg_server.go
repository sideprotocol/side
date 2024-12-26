package keeper

import (
	"context"

	"github.com/sideprotocol/side/x/lending/types"
)

type msgServer struct {
	Keeper
}

// CreateLoan implements types.MsgServer.
func (m msgServer) Apply(ctx context.Context, msg *types.MsgApply) (*types.MsgApplyResponse, error) {
	panic("unimplemented")
}

// Repay implements types.MsgServer.
func (m msgServer) Repay(ctx context.Context, msg *types.MsgRepay) (*types.MsgRepayResponse, error) {
	panic("unimplemented")
}

// RequestVaultAddress implements types.MsgServer.
func (m msgServer) Redeem(ctx context.Context, msg *types.MsgRedeem) (*types.MsgRedeemResponse, error) {
	panic("unimplemented")
}

// SubmitFundingTx implements types.MsgServer.
func (m msgServer) Fund(ctx context.Context, msg *types.MsgFund) (*types.MsgFundResponse, error) {
	panic("unimplemented")
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
