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

// CreateLoan implements types.MsgServer.
func (m msgServer) CreateLoan(context.Context, *types.MsgCreateLoan) (*types.MsgCreateLoanResponse, error) {
	panic("unimplemented")
}

// RemoveLiquidity implements types.MsgServer.
func (m msgServer) RemoveLiquidity(context.Context, *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	panic("unimplemented")
}

// Repay implements types.MsgServer.
func (m msgServer) Repay(context.Context, *types.MsgRepay) (*types.MsgRepayResponse, error) {
	panic("unimplemented")
}

// RequestVaultAddress implements types.MsgServer.
func (m msgServer) RequestVaultAddress(context.Context, *types.MsgRequestVaultAddress) (*types.MsgRequestVaultAddressResponse, error) {
	panic("unimplemented")
}

// SubmitFundingTx implements types.MsgServer.
func (m msgServer) SubmitFundingTx(context.Context, *types.MsgSubmitFundingTx) (*types.MsgSubmitFundingTxResponse, error) {
	panic("unimplemented")
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
