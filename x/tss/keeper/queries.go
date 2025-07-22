package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/sideprotocol/side/x/tss/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) DKGRequest(goCtx context.Context, req *types.QueryDKGRequestRequest) (*types.QueryDKGRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasDKGRequest(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "dkg request does not exist")
	}

	return &types.QueryDKGRequestResponse{Request: k.GetDKGRequest(ctx, req.Id)}, nil
}

func (k Keeper) DKGRequests(goCtx context.Context, req *types.QueryDKGRequestsRequest) (*types.QueryDKGRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var err error
	var dkgRequests []*types.DKGRequest
	var pagination *query.PageResponse

	if req.Status == types.DKGStatus_DKG_STATUS_UNSPECIFIED {
		dkgRequests, pagination, err = k.GetDKGRequestsWithPagination(ctx, req.Module, req.Pagination)
	} else {
		dkgRequests, pagination, err = k.GetDKGRequestsByStatusWithPagination(ctx, req.Status, req.Module, req.Pagination)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDKGRequestsResponse{Requests: dkgRequests, Pagination: pagination}, nil
}

func (k Keeper) DKGCompletions(goCtx context.Context, req *types.QueryDKGCompletionsRequest) (*types.QueryDKGCompletionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryDKGCompletionsResponse{Completions: k.GetDKGCompletions(ctx, req.Id)}, nil
}

func (k Keeper) SigningRequest(goCtx context.Context, req *types.QuerySigningRequestRequest) (*types.QuerySigningRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasSigningRequest(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "signing request does not exist")
	}

	return &types.QuerySigningRequestResponse{Request: k.GetSigningRequest(ctx, req.Id)}, nil
}

func (k Keeper) SigningRequests(goCtx context.Context, req *types.QuerySigningRequestsRequest) (*types.QuerySigningRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var err error
	var signingRequests []*types.SigningRequest
	var pagination *query.PageResponse

	if req.Status == types.SigningStatus_SIGNING_STATUS_UNSPECIFIED {
		signingRequests, pagination, err = k.GetSigningRequestsWithPagination(ctx, req.Module, req.Pagination)
	} else {
		signingRequests, pagination, err = k.GetSigningRequestsByStatusWithPagination(ctx, req.Status, req.Module, req.Pagination)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySigningRequestsResponse{Requests: signingRequests, Pagination: pagination}, nil
}

func (k Keeper) RefreshingRequest(goCtx context.Context, req *types.QueryRefreshingRequestRequest) (*types.QueryRefreshingRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasRefreshingRequest(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "refreshing request does not exist")
	}

	return &types.QueryRefreshingRequestResponse{Request: k.GetRefreshingRequest(ctx, req.Id)}, nil
}

func (k Keeper) RefreshingRequests(goCtx context.Context, req *types.QueryRefreshingRequestsRequest) (*types.QueryRefreshingRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryRefreshingRequestsResponse{Requests: k.GetRefreshingRequests(ctx, req.Status)}, nil
}

func (k Keeper) RefreshingCompletions(goCtx context.Context, req *types.QueryRefreshingCompletionsRequest) (*types.QueryRefreshingCompletionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryRefreshingCompletionsResponse{Completions: k.GetRefreshingCompletions(ctx, req.Id)}, nil
}
