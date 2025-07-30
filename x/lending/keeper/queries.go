package keeper

import (
	"context"
	"encoding/hex"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

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

	return &types.QueryPoolResponse{Pool: pool}, nil
}

// Pools implements types.QueryServer.
func (k Keeper) Pools(goCtx context.Context, req *types.QueryPoolsRequest) (*types.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryPoolsResponse{Pools: k.GetAllPools(ctx)}, nil
}

// PoolExchangeRate implements types.QueryServer.
func (k Keeper) PoolExchangeRate(goCtx context.Context, req *types.QueryPoolExchangeRateRequest) (*types.QueryPoolExchangeRateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasPool(ctx, req.PoolId) {
		return nil, status.Error(codes.InvalidArgument, "pool does not exist")
	}

	pool := k.GetPool(ctx, req.PoolId)

	exchangeRate := types.GetExchangeRate(pool.AvailableAmount, pool.TotalBorrowed, pool.TotalReserve, pool.TotalYTokens.Amount)

	return &types.QueryPoolExchangeRateResponse{ExchangeRate: exchangeRate.String()}, nil
}

// CollateralAddress implements types.QueryServer.
func (k Keeper) CollateralAddress(goCtx context.Context, req *types.QueryCollateralAddressRequest) (*types.QueryCollateralAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	borrowerPubKey, err := hex.DecodeString(req.BorrowerPubkey)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to decode borrower pub key")
	}

	if _, err := schnorr.ParsePubKey(borrowerPubKey); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid borrower pub key")
	}

	borrowerAuthPubKey, err := hex.DecodeString(req.BorrowerAuthPubkey)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to decode borrower auth pub key")
	}

	if _, err := schnorr.ParsePubKey(borrowerAuthPubKey); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid borrower auth pub key")
	}

	dcmPubKey, err := hex.DecodeString(req.DCMPubKey)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to decode dcm pub key")
	}

	if _, err := schnorr.ParsePubKey(dcmPubKey); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid dcm pub key")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	collateralAddr, err := types.CreateVaultAddress(req.BorrowerPubkey, req.BorrowerAuthPubkey, req.DCMPubKey, int64(req.MaturityTime)+k.FinalTimeoutDuration(ctx))
	if err != nil {
		return nil, err
	}

	return &types.QueryCollateralAddressResponse{Address: collateralAddr}, nil
}

// LiquidationPrice implements types.QueryServer.
func (k Keeper) LiquidationPrice(goCtx context.Context, req *types.QueryLiquidationPriceRequest) (*types.QueryLiquidationPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasPool(ctx, req.PoolId) {
		return nil, status.Error(codes.InvalidArgument, "pool does not exist")
	}

	poolConfig := k.GetPool(ctx, req.PoolId).Config

	trancheConfig, found := types.GetTrancheConfig(poolConfig.Tranches, req.Maturity)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "maturity does not exit")
	}

	collateralAmount, err := sdk.ParseCoinNormalized(req.CollateralAmount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	borrowedAmount, err := sdk.ParseCoinNormalized(req.BorrowAmount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	liquidationPrice := types.GetLiquidationPrice(collateralAmount.Amount, int(poolConfig.CollateralAsset.Decimals), borrowedAmount.Amount, int(poolConfig.LendingAsset.Decimals), trancheConfig.Maturity, trancheConfig.BorrowAPR, k.GetBlocksPerYear(ctx), poolConfig.LiquidationThreshold, poolConfig.CollateralAsset.IsBasePriceAsset)

	return &types.QueryLiquidationPriceResponse{
		Price: types.FormatPrice(liquidationPrice),
		Pair:  types.GetPricePair(poolConfig),
	}, nil
}

// DlcEventCount implements types.QueryServer.
func (k Keeper) DlcEventCount(goCtx context.Context, req *types.QueryDlcEventCountRequest) (*types.QueryDlcEventCountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryDlcEventCountResponse{Count: k.dlcKeeper.GetPendingLendingEventCount(ctx)}, nil
}

// Loan implements types.QueryServer.
func (k Keeper) Loan(goCtx context.Context, req *types.QueryLoanRequest) (*types.QueryLoanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLoan(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "loan does not exist")
	}

	loan := k.GetLoan(ctx, req.Id)

	return &types.QueryLoanResponse{Loan: loan}, nil
}

// Loans implements types.QueryServer.
func (k Keeper) Loans(goCtx context.Context, req *types.QueryLoansRequest) (*types.QueryLoansResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var err error
	var loans []*types.Loan
	var pagination *query.PageResponse

	if req.Status == types.LoanStatus_Unspecified {
		loans, pagination, err = k.GetLoansWithPagination(ctx, req.Pagination)
	} else {
		loans, pagination, err = k.GetLoansByStatusWithPagination(ctx, req.Status, req.Pagination)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryLoansResponse{Loans: loans, Pagination: pagination}, nil
}

// LoansByAddress implements types.QueryServer.
func (k Keeper) LoansByAddress(goCtx context.Context, req *types.QueryLoansByAddressRequest) (*types.QueryLoansByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	loans, pagination, err := k.GetLoansByAddress(ctx, req.Address, req.Status, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryLoansByAddressResponse{Loans: loans, Pagination: pagination}, nil
}

// LoansByOracle implements types.QueryServer.
func (k Keeper) LoansByOracle(goCtx context.Context, req *types.QueryLoansByOracleRequest) (*types.QueryLoansByOracleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	oraclePubKey, err := hex.DecodeString(req.OraclePubkey)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid oracle pub key")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	loans, pagination, err := k.GetLoansByOracle(ctx, oraclePubKey, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryLoansByOracleResponse{Loans: loans, Pagination: pagination}, nil
}

// LoanCetInfos implements types.QueryServer.
func (k Keeper) LoanCetInfos(goCtx context.Context, req *types.QueryLoanCetInfosRequest) (*types.QueryLoanCetInfosResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLoan(ctx, req.LoanId) {
		return nil, status.Error(codes.InvalidArgument, "loan does not exist")
	}

	loan := k.GetLoan(ctx, req.LoanId)

	var err error
	var collateralAmount sdk.Coin

	if loan.LiquidationPrice.IsZero() {
		collateralAmount, err = sdk.ParseCoinNormalized(req.CollateralAmount)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if !collateralAmount.IsPositive() {
			return nil, status.Error(codes.InvalidArgument, "collateral amount must be positive")
		}
	}

	cetInfos, err := k.GetCetInfos(ctx, req.LoanId, collateralAmount)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryLoanCetInfosResponse{
		LiquidationCetInfo:        cetInfos[0],
		DefaultLiquidationCetInfo: cetInfos[1],
		RepaymentCetInfo:          cetInfos[2],
	}, nil
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

// LoanAuthorization implements types.QueryServer.
func (k Keeper) LoanAuthorization(goCtx context.Context, req *types.QueryLoanAuthorizationRequest) (*types.QueryLoanAuthorizationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLoan(ctx, req.LoanId) {
		return nil, status.Error(codes.InvalidArgument, "loan does not exist")
	}

	if !k.HasAuthorization(ctx, req.LoanId, req.Id) {
		return nil, status.Error(codes.NotFound, "loan authorization does not exist")
	}

	authorization := k.GetAuthorization(ctx, req.LoanId, req.Id)

	return &types.QueryLoanAuthorizationResponse{
		Deposits: k.GetDeposits(ctx, authorization),
		Status:   authorization.Status,
	}, nil
}

// LoanDeposits implements types.QueryServer.
func (k Keeper) LoanDeposits(goCtx context.Context, req *types.QueryLoanDepositsRequest) (*types.QueryLoanDepositsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLoan(ctx, req.LoanId) {
		return nil, status.Error(codes.InvalidArgument, "loan does not exist")
	}

	return &types.QueryLoanDepositsResponse{Deposits: k.GetDepositLogs(ctx, req.LoanId)}, nil
}

// Redemption implements types.QueryServer.
func (k Keeper) Redemption(goCtx context.Context, req *types.QueryRedemptionRequest) (*types.QueryRedemptionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasRedemption(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "redemption does not exist")
	}

	return &types.QueryRedemptionResponse{Redemption: k.GetRedemption(ctx, req.Id)}, nil
}

// Repayment implements types.QueryServer.
func (k Keeper) Repayment(goCtx context.Context, req *types.QueryRepaymentRequest) (*types.QueryRepaymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLoan(ctx, req.LoanId) {
		return nil, status.Error(codes.InvalidArgument, "loan does not exist")
	}

	if !k.HasRepayment(ctx, req.LoanId) {
		return nil, status.Error(codes.NotFound, "repayment does not exist")
	}

	return &types.QueryRepaymentResponse{Repayment: k.GetRepayment(ctx, req.LoanId)}, nil
}

// CurrentInterest implements types.QueryServer.
func (k Keeper) CurrentInterest(goCtx context.Context, req *types.QueryCurrentInterestRequest) (*types.QueryCurrentInterestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLoan(ctx, req.LoanId) {
		return nil, status.Error(codes.InvalidArgument, "loan does not exist")
	}

	currentInterest := k.GetCurrentInterest(ctx, k.GetLoan(ctx, req.LoanId))

	return &types.QueryCurrentInterestResponse{
		Interest: currentInterest,
	}, nil
}

// Referrer implements types.QueryServer.
func (k Keeper) Referrer(goCtx context.Context, req *types.QueryReferrerRequest) (*types.QueryReferrerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasReferrer(ctx, req.ReferralCode) {
		return nil, status.Error(codes.NotFound, "referrer does not exist")
	}

	return &types.QueryReferrerResponse{Referrer: k.GetReferrer(ctx, req.ReferralCode)}, nil
}

// Referrers implements types.QueryServer.
func (k Keeper) Referrers(goCtx context.Context, req *types.QueryReferrersRequest) (*types.QueryReferrersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	referrers, pagination, err := k.GetReferrersWithPagination(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryReferrersResponse{Referrers: referrers, Pagination: pagination}, nil
}

// Params implements types.QueryServer.
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}
