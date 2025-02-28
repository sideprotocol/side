package keeper

import (
	"context"
	"encoding/hex"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	"github.com/sideprotocol/side/x/lending/types"
)

var _ types.QueryServer = Keeper{}

// Pool implements types.QueryServer.
func (k Keeper) Pool(goCtx context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasPool(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "pool does not exist")
	}

	pool := k.GetPool(ctx, req.Id)

	return &types.QueryPoolResponse{Pool: &pool}, nil
}

// Pools implements types.QueryServer.
func (k Keeper) Pools(goCtx context.Context, req *types.QueryPoolsRequest) (*types.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryPoolsResponse{Pools: k.GetAllPools(ctx)}, nil
}

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

	collateralAmount, err := sdk.ParseCoinNormalized(req.CollateralAmount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	borrowedAmount, err := sdk.ParseCoinNormalized(req.BorrowAmount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	liquidationPrice := types.GetLiquidationPrice(collateralAmount.Amount, borrowedAmount.Amount, k.GetParams(ctx).LiquidationThresholdPercent)

	event := k.dlcKeeper.GetEventByPrice(ctx, liquidationPrice)
	if event == nil {
		return nil, status.Error(codes.NotFound, "liquidation event does not exist")
	}

	signaturePoint, err := dlctypes.GetSignaturePointFromEvent(event)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryLiquidationEventResponse{
		EventId:        event.Id,
		OraclePubkey:   event.Pubkey,
		Nonce:          event.Nonce,
		Price:          event.TriggerPrice.String(),
		SignaturePoint: hex.EncodeToString(signaturePoint),
	}, nil
}

// LiquidationCet implements types.QueryServer.
func (k Keeper) LiquidationCet(goCtx context.Context, req *types.QueryLiquidationCetRequest) (*types.QueryLiquidationCetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var err error
	var script string
	var sigHashes []string

	if len(req.LoanId) != 0 {
		if !k.HasLoan(ctx, req.LoanId) {
			return nil, status.Error(codes.InvalidArgument, "loan does not exist")
		}

		dlcMeta := k.GetDLCMeta(ctx, req.LoanId)
		script = dlcMeta.LiquidationCetScript

		sigHashes, err = types.GetLiquidationCetSigHashes(dlcMeta)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		scriptBytes, err := types.CreateMultisigScript([]string{req.BorrowerPubkey, req.AgencyPubkey})
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		script = hex.EncodeToString(scriptBytes)
	}

	return &types.QueryLiquidationCetResponse{
		Script:    script,
		SigHashes: sigHashes,
	}, nil
}

// Loan implements types.QueryServer.
func (k Keeper) Loan(goCtx context.Context, req *types.QueryLoanRequest) (*types.QueryLoanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLoan(ctx, req.LoanId) {
		return nil, status.Error(codes.NotFound, "loan does not exist")
	}

	loan := k.GetLoan(ctx, req.LoanId)

	return &types.QueryLoanResponse{Loan: &loan}, nil
}

// Loans implements types.QueryServer.
func (k Keeper) Loans(goCtx context.Context, req *types.QueryLoansRequest) (*types.QueryLoansResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var loans []*types.Loan

	if req.Status == types.LoanStatus_Unspecified {
		loans = k.GetAllLoans(ctx)
	} else {
		loans = k.GetLoans(ctx, req.Status)
	}

	return &types.QueryLoansResponse{Loans: loans}, nil
}

// LoanDlcMeta implements types.QueryServer.
func (k Keeper) LoanDlcMeta(goCtx context.Context, req *types.QueryLoanDlcMetaRequest) (*types.QueryLoanDlcMetaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLoan(ctx, req.LoanId) {
		return nil, status.Error(codes.InvalidArgument, "loan does not exist")
	}

	return &types.QueryLoanDlcMetaResponse{DlcMeta: k.GetDLCMeta(ctx, req.LoanId)}, nil
}

// Repayment implements types.QueryServer.
func (k Keeper) Repayment(goCtx context.Context, req *types.QueryRepaymentRequest) (*types.QueryRepaymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasRepayment(ctx, req.LoanId) {
		return nil, status.Error(codes.NotFound, "repayment does not exist")
	}

	repayment := k.GetRepayment(ctx, req.LoanId)

	return &types.QueryRepaymentResponse{Repayment: &repayment}, nil
}

// Params implements types.QueryServer.
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}
