package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) QueryParams(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) QueryFeeRate(goCtx context.Context, req *types.QueryFeeRateRequest) (*types.QueryFeeRateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	feeRate := k.GetFeeRate(ctx)

	return &types.QueryFeeRateResponse{FeeRate: feeRate}, nil
}

func (k Keeper) QueryWithdrawalNetworkFee(goCtx context.Context, req *types.QueryWithdrawalNetworkFeeRequest) (*types.QueryWithdrawalNetworkFeeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	amount, err := sdk.ParseCoinNormalized(req.Amount)
	if err != nil {
		return nil, err
	}

	feeRate := req.FeeRate
	if feeRate < 0 {
		return nil, types.ErrInvalidFeeRate
	}

	if feeRate == 0 {
		currentFeeRate := k.GetFeeRate(ctx)
		if err := k.CheckFeeRate(ctx, currentFeeRate); err != nil {
			return nil, err
		}

		feeRate = currentFeeRate.Value
	}

	networkFee, err := k.EstimateWithdrawalNetworkFee(ctx, req.Address, amount, feeRate)
	if err != nil {
		return nil, err
	}

	return &types.QueryWithdrawalNetworkFeeResponse{FeeRate: feeRate, Fee: networkFee.String()}, nil
}

func (k Keeper) QueryWithdrawRequestsByAddress(goCtx context.Context, req *types.QueryWithdrawRequestsByAddressRequest) (*types.QueryWithdrawRequestsByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	requests := k.GetWithdrawRequestsByAddress(ctx, req.Address)

	return &types.QueryWithdrawRequestsByAddressResponse{Requests: requests}, nil
}

func (k Keeper) QueryWithdrawRequestsByTxHash(goCtx context.Context, req *types.QueryWithdrawRequestsByTxHashRequest) (*types.QueryWithdrawRequestsByTxHashResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests := k.GetWithdrawRequestsByTxHash(ctx, req.Txid)

	return &types.QueryWithdrawRequestsByTxHashResponse{Requests: requests}, nil
}

func (k Keeper) QueryPendingBtcWithdrawRequests(goCtx context.Context, req *types.QueryPendingBtcWithdrawRequestsRequest) (*types.QueryPendingBtcWithdrawRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests := k.GetPendingBtcWithdrawRequests(ctx, 0)

	return &types.QueryPendingBtcWithdrawRequestsResponse{Requests: requests}, nil
}

func (k Keeper) QuerySigningRequest(goCtx context.Context, req *types.QuerySigningRequestRequest) (*types.QuerySigningRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	request := k.GetSigningRequest(ctx, req.Sequence)

	return &types.QuerySigningRequestResponse{Request: request}, nil
}

func (k Keeper) QuerySigningRequests(goCtx context.Context, req *types.QuerySigningRequestsRequest) (*types.QuerySigningRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Status == types.SigningStatus_SIGNING_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests, pagination, err := k.GetSigningRequestsByStatus(ctx, req.Status, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySigningRequestsResponse{Requests: requests, Pagination: pagination}, nil
}

func (k Keeper) QuerySigningRequestsByAddress(goCtx context.Context, req *types.QuerySigningRequestsByAddressRequest) (*types.QuerySigningRequestsByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests := k.GetSigningRequestsByAddr(ctx, req)

	return &types.QuerySigningRequestsByAddressResponse{Requests: requests}, nil
}

func (k Keeper) QuerySigningRequestByTxHash(goCtx context.Context, req *types.QuerySigningRequestByTxHashRequest) (*types.QuerySigningRequestByTxHashResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var request *types.SigningRequest

	if k.HasSigningRequestByTxHash(ctx, req.Txid) {
		request = k.GetSigningRequestByTxHash(ctx, req.Txid)
	}

	return &types.QuerySigningRequestByTxHashResponse{Request: request}, nil
}

func (k Keeper) QueryPendingSigningRequests(goCtx context.Context, req *types.QueryPendingSigningRequestsRequest) (*types.QueryPendingSigningRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests, pagination, err := k.GetCompactSigningRequestsByStatus(ctx, types.SigningStatus_SIGNING_STATUS_PENDING, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPendingSigningRequestsResponse{Requests: requests, Pagination: pagination}, nil
}

func (k Keeper) QueryUTXOs(goCtx context.Context, req *types.QueryUTXOsRequest) (*types.QueryUTXOsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	utxos := k.GetAllUTXOs(ctx)

	return &types.QueryUTXOsResponse{Utxos: utxos}, nil
}

func (k Keeper) QueryUTXOsByAddress(goCtx context.Context, req *types.QueryUTXOsByAddressRequest) (*types.QueryUTXOsByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	utxos := k.GetUTXOsByAddr(ctx, req.Address)

	return &types.QueryUTXOsByAddressResponse{Utxos: utxos}, nil
}

func (k Keeper) QueryUTXOCountAndBalancesByAddress(goCtx context.Context, req *types.QueryUTXOCountAndBalancesByAddressRequest) (*types.QueryUTXOCountAndBalancesByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	count, value, runeBalances := k.GetUnlockedUTXOCountAndBalancesByAddr(ctx, req.Address)

	return &types.QueryUTXOCountAndBalancesByAddressResponse{
		Count:        count,
		Value:        value,
		RuneBalances: runeBalances,
	}, nil
}

func (k Keeper) QueryDKGRequest(goCtx context.Context, req *types.QueryDKGRequestRequest) (*types.QueryDKGRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	request := k.GetDKGRequest(ctx, req.Id)

	return &types.QueryDKGRequestResponse{Request: request}, nil
}

func (k Keeper) QueryDKGRequests(goCtx context.Context, req *types.QueryDKGRequestsRequest) (*types.QueryDKGRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests := k.GetDKGRequests(ctx, req.Status)

	return &types.QueryDKGRequestsResponse{Requests: requests}, nil
}

func (k Keeper) QueryAllDKGRequests(goCtx context.Context, req *types.QueryAllDKGRequestsRequest) (*types.QueryAllDKGRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests := k.GetAllDKGRequests(ctx)

	return &types.QueryAllDKGRequestsResponse{Requests: requests}, nil
}

func (k Keeper) QueryDKGCompletionRequests(goCtx context.Context, req *types.QueryDKGCompletionRequestsRequest) (*types.QueryDKGCompletionRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests := k.GetDKGCompletionRequests(ctx, req.Id)

	return &types.QueryDKGCompletionRequestsResponse{Requests: requests}, nil
}
