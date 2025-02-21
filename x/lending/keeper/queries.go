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

	return &types.QueryLoansResponse{Loans: k.GetLoans(ctx, req.Status)}, nil
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
