package keeper

import (
	"context"

	"github.com/sideprotocol/side/x/lending/types"
)

var _ types.QueryServer = Keeper{}

// CollateralAddress implements types.QueryServer.
func (k Keeper) CollateralAddress(context.Context, *types.QueryCollateralAddressRequest) (*types.QueryCollateralAddressResponse, error) {
	panic("unimplemented")
}

// LiquidationEvent implements types.QueryServer.
func (k Keeper) LiquidationEvent(context.Context, *types.QueryLiquidationEventRequest) (*types.QueryLiquidationEventResponse, error) {
	panic("unimplemented")
}

// LoanCETs implements types.QueryServer.
func (k Keeper) LoanCETs(context.Context, *types.QueryLoanCETsRequest) (*types.QueryLoanCETsResponse, error) {
	panic("unimplemented")
}

// Params implements types.QueryServer.
func (k Keeper) Params(context.Context, *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	panic("unimplemented")
}

// UnsignedPaymentTx implements types.QueryServer.
func (k Keeper) UnsignedPaymentTx(context.Context, *types.QueryRepaymentTxRequest) (*types.QueryRepaymentTxResponse, error) {
	panic("unimplemented")
}
