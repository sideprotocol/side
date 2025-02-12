package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

var _ types.QueryServer = Keeper{}

// CollateralAddress implements types.QueryServer.
func (k Keeper) CollateralAddress(goCtx context.Context, req *types.QueryCollateralAddressRequest) (*types.QueryCollateralAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	collateralAddr, err := types.CreateVaultAddress(req.BorrowerPubkey, req.AgencyPubkey, req.HashOfLoanSecret, int64(req.MaturityTime), int64(req.FinalTimeout))
	if err != nil {
		return nil, err
	}

	return &types.QueryCollateralAddressResponse{Address: collateralAddr}, nil
}

// LiquidationEvent implements types.QueryServer.
func (k Keeper) LiquidationEvent(goCtx context.Context, req *types.QueryLiquidationEventRequest) (*types.QueryLiquidationEventResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	liquidationPrice := types.GetLiquidationPrice(req.CollateralAcmount.Amount, req.BorrowAmount.Amount, k.GetParams(ctx).LiquidationThresholdPercent)

	event := k.dlcKeeper.GetEventByPrice(ctx, liquidationPrice)
	if event == nil {
		return nil, nil
	}

	return &types.QueryLiquidationEventResponse{
		EventId:      event.Id,
		OraclePubkey: event.Pubkey,
		Nonce:        event.Nonce,
		Price:        event.TriggerPrice.String(),
	}, nil
}

// LoanCETs implements types.QueryServer.
func (k Keeper) LoanCETs(goCtx context.Context, req *types.QueryLoanCETsRequest) (*types.QueryLoanCETsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryLoanCETsResponse{CETs: k.GetCETs(ctx, req.LoanId)}, nil
}

// UnsignedPaymentTx implements types.QueryServer.
func (k Keeper) UnsignedPaymentTx(goCtx context.Context, req *types.QueryRepaymentTxRequest) (*types.QueryRepaymentTxResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryRepaymentTxResponse{ClaimTx: k.GetRepayment(ctx, req.LoanId).Tx}, nil
}

// Params implements types.QueryServer.
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}
