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

func (k Keeper) DCMs(goCtx context.Context, req *types.QueryDCMsRequest) (*types.QueryDCMsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryDCMsResponse{DCMs: k.GetDCMs(ctx, req.Status)}, nil
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

	events, pagination, err := k.GetEventsByStatusWithPagination(ctx, req.Triggered, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryEventsResponse{Events: events, Pagination: pagination}, nil
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

func (k Keeper) AttestationByEvent(goCtx context.Context, req *types.QueryAttestationByEventRequest) (*types.QueryAttestationByEventResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasAttestationByEvent(ctx, req.EventId) {
		return nil, status.Error(codes.NotFound, "attestation does not exist")
	}

	return &types.QueryAttestationByEventResponse{Attestation: k.GetAttestationByEvent(ctx, req.EventId)}, nil
}

func (k Keeper) Attestations(goCtx context.Context, req *types.QueryAttestationsRequest) (*types.QueryAttestationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryAttestationsResponse{Attestations: k.GetAttestations(ctx)}, nil
}

func (k Keeper) OracleParticipantLiveness(goCtx context.Context, req *types.QueryOracleParticipantLivenessRequest) (*types.QueryOracleParticipantLivenessResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	participantsLiveness := []*types.OracleParticipantLiveness{}

	if len(req.ConsensusPubkey) != 0 {
		if !k.HasOracleParticipantLiveness(ctx, req.ConsensusPubkey) {
			return nil, status.Error(codes.NotFound, "oracle participant liveness does not exist")
		}

		participantsLiveness = append(participantsLiveness, k.GetOracleParticipantLiveness(ctx, req.ConsensusPubkey))
	} else {
		participantsLiveness = k.GetOracleParticipantsLiveness(ctx, req.Alive)
	}

	return &types.QueryOracleParticipantLivenessResponse{ParticipantLivenesses: participantsLiveness}, nil
}
