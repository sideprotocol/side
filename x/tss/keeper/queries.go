package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	return &types.QueryDKGRequestsResponse{Requests: k.GetDKGRequests(ctx, req.Status)}, nil
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

	return &types.QuerySigningRequestsResponse{Requests: k.GetSigningRequests(ctx, req.Status)}, nil
}

func (k Keeper) ResharingRequest(goCtx context.Context, req *types.QueryResharingRequestRequest) (*types.QueryResharingRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasResharingRequest(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "resharing request does not exist")
	}

	return &types.QueryResharingRequestResponse{Request: k.GetResharingRequest(ctx, req.Id)}, nil
}

func (k Keeper) ResharingRequests(goCtx context.Context, req *types.QueryResharingRequestsRequest) (*types.QueryResharingRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryResharingRequestsResponse{Requests: k.GetResharingRequests(ctx, req.Status)}, nil
}

func (k Keeper) ResharingCompletions(goCtx context.Context, req *types.QueryResharingCompletionsRequest) (*types.QueryResharingCompletionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryResharingCompletionsResponse{Completions: k.GetResharingCompletions(ctx, req.Id)}, nil
}
