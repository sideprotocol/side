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

func (k Keeper) QueryChainTip(goCtx context.Context, req *types.QueryChainTipRequest) (*types.QueryChainTipResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	best := k.GetBestBlockHeader(ctx)

	return &types.QueryChainTipResponse{
		Hash:   best.Hash,
		Height: best.Height,
	}, nil
}

// BlockHeaderByHash queries the block header by hash.
func (k Keeper) QueryBlockHeaderByHash(goCtx context.Context, req *types.QueryBlockHeaderByHashRequest) (*types.QueryBlockHeaderByHashResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	header := k.GetBlockHeader(ctx, req.Hash)
	if header == nil {
		return nil, status.Error(codes.NotFound, "block header not found")
	}

	return &types.QueryBlockHeaderByHashResponse{BlockHeader: header}, nil
}

func (k Keeper) QueryBlockHeaderByHeight(goCtx context.Context, req *types.QueryBlockHeaderByHeightRequest) (*types.QueryBlockHeaderByHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	header := k.GetBlockHeaderByHeight(ctx, req.Height)
	if header == nil {
		return nil, status.Error(codes.NotFound, "block header not found")
	}

	return &types.QueryBlockHeaderByHeightResponse{BlockHeader: header}, nil
}

func (k Keeper) QueryFeeRate(goCtx context.Context, req *types.QueryFeeRateRequest) (*types.QueryFeeRateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	feeRate := k.GetFeeRate(ctx)

	return &types.QueryFeeRateResponse{FeeRate: feeRate}, nil
}

func (k Keeper) QuerySigningRequests(goCtx context.Context, req *types.QuerySigningRequestsRequest) (*types.QuerySigningRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Status == types.SigningStatus_SIGNING_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests := k.FilterSigningRequestsByStatus(ctx, req)

	return &types.QuerySigningRequestsResponse{Requests: requests}, nil
}

func (k Keeper) QuerySigningRequestsByAddress(goCtx context.Context, req *types.QuerySigningRequestsByAddressRequest) (*types.QuerySigningRequestsByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests := k.FilterSigningRequestsByAddr(ctx, req)

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
