package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/btcbridge/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (k Keeper) QueryWithdrawRequests(goCtx context.Context, req *types.QueryWithdrawRequestsRequest) (*types.QueryWithdrawRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Status == types.WithdrawStatus_WITHDRAW_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests := k.FilterWithdrawRequestsByStatus(ctx, req)

	return &types.QueryWithdrawRequestsResponse{Requests: requests}, nil
}

func (k Keeper) QueryWithdrawRequestsByAddress(goCtx context.Context, req *types.QueryWithdrawRequestsByAddressRequest) (*types.QueryWithdrawRequestsByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests := k.FilterWithdrawRequestsByAddr(ctx, req)

	return &types.QueryWithdrawRequestsByAddressResponse{Requests: requests}, nil
}

func (k Keeper) QueryWithdrawRequestByTxHash(goCtx context.Context, req *types.QueryWithdrawRequestByTxHashRequest) (*types.QueryWithdrawRequestByTxHashResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var request *types.BitcoinWithdrawRequest

	if k.HasWithdrawRequest(ctx, req.Txid) {
		request = k.GetWithdrawRequest(ctx, req.Txid)
	}

	return &types.QueryWithdrawRequestByTxHashResponse{Request: request}, nil
}
