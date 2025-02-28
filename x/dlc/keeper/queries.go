package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) Oracles(goCtx context.Context, req *types.QueryOraclesRequest) (*types.QueryOraclesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryOraclesResponse{Oracles: k.GetOracles(ctx, req.Status)}, nil
}

func (k Keeper) Agencies(goCtx context.Context, req *types.QueryAgenciesRequest) (*types.QueryAgenciesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryAgenciesResponse{Agencies: k.GetAgencies(ctx, req.Status)}, nil
}

func (k Keeper) Nonce(goCtx context.Context, req *types.QueryNonceRequest) (*types.QueryNonceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryNonceResponse{Nonce: k.GetNonce(ctx, req.OracleId, req.Index)}, nil
}

func (k Keeper) Nonces(goCtx context.Context, req *types.QueryNoncesRequest) (*types.QueryNoncesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryNoncesResponse{Nonces: k.GetNonces(ctx, req.OracleId)}, nil
}

func (k Keeper) CountNonces(goCtx context.Context, req *types.QueryCountNoncesRequest) (*types.QueryCountNoncesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryCountNoncesResponse{Counts: k.GetNonceCounts(ctx)}, nil
}

func (k Keeper) Event(goCtx context.Context, req *types.QueryEventRequest) (*types.QueryEventResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasEvent(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "event does not exist")
	}

	return &types.QueryEventResponse{Event: k.GetEvent(ctx, req.Id)}, nil
}

func (k Keeper) Events(goCtx context.Context, req *types.QueryEventsRequest) (*types.QueryEventsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryEventsResponse{Events: k.GetEvents(ctx, req.Triggered)}, nil
}

func (k Keeper) Attestation(goCtx context.Context, req *types.QueryAttestationRequest) (*types.QueryAttestationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasAttestation(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "attestation does not exist")
	}

	return &types.QueryAttestationResponse{Attestation: k.GetAttestation(ctx, req.Id)}, nil
}

func (k Keeper) Attestations(goCtx context.Context, req *types.QueryAttestationsRequest) (*types.QueryAttestationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryAttestationsResponse{Attestations: k.GetAttestations(ctx)}, nil
}

func (k Keeper) Price(goCtx context.Context, req *types.QueryPriceRequest) (*types.QueryPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryPriceResponse{Price: k.GetPrice(ctx, req.Symbol).Uint64()}, nil
}
