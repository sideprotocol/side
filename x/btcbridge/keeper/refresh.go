package keeper

import (
	"fmt"
	"slices"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// GetRefreshingRequestId returns the refreshing request id
func (k Keeper) GetRefreshingRequestId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.RefreshingRequestIdKey)

	return sdk.BigEndianToUint64(bz)
}

// IncrementRefreshingRequestId increments the refreshing request id and returns the new id
func (k Keeper) IncrementRefreshingRequestId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetRefreshingRequestId(ctx) + 1
	store.Set(types.RefreshingRequestIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// SetRefreshingRequest sets the refreshing request
func (k Keeper) SetRefreshingRequest(ctx sdk.Context, refreshingRequest *types.RefreshingRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(refreshingRequest)

	store.Set(types.RefreshingRequestKey(refreshingRequest.Id), bz)
}

// HasRefreshingRequest returns true if the given refreshing request exists, false otherwise
func (k Keeper) HasRefreshingRequest(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.RefreshingRequestKey(id))
}

// GetRefreshingRequest returns the refreshing request by the given id
func (k Keeper) GetRefreshingRequest(ctx sdk.Context, id uint64) *types.RefreshingRequest {
	store := ctx.KVStore(k.storeKey)

	var refreshingRequest types.RefreshingRequest
	bz := store.Get(types.RefreshingRequestKey(id))
	k.cdc.MustUnmarshal(bz, &refreshingRequest)

	return &refreshingRequest
}

// GetRefreshingParticipants gets all participants of the given refreshing request
func (k Keeper) GetRefreshingParticipants(ctx sdk.Context, refreshingRequest *types.RefreshingRequest) []*types.DKGParticipant {
	dkgReq := k.GetDKGRequest(ctx, refreshingRequest.DkgId)

	participants := []*types.DKGParticipant{}
	for _, p := range dkgReq.Participants {
		if !slices.Contains(refreshingRequest.RemovedParticipants, p.ConsensusPubkey) {
			participants = append(participants, p)
		}
	}

	return participants
}

// GetRefreshingRequests gets the refreshing requests by the given status
func (k Keeper) GetRefreshingRequests(ctx sdk.Context, status types.RefreshingStatus) []*types.RefreshingRequest {
	requests := make([]*types.RefreshingRequest, 0)

	k.IterateRefreshingRequests(ctx, func(req *types.RefreshingRequest) (stop bool) {
		if req.Status == status {
			requests = append(requests, req)
		}

		return false
	})

	return requests
}

// GetPendingRefreshingRequests gets the pending refreshing requests
func (k Keeper) GetPendingRefreshingRequests(ctx sdk.Context) []*types.RefreshingRequest {
	requests := make([]*types.RefreshingRequest, 0)

	k.IterateRefreshingRequests(ctx, func(req *types.RefreshingRequest) (stop bool) {
		if req.Status == types.RefreshingStatus_REFRESHING_STATUS_PENDING {
			requests = append(requests, req)
		}

		return false
	})

	return requests
}

// IterateRefreshingRequests iterates through all refreshing requests
func (k Keeper) IterateRefreshingRequests(ctx sdk.Context, cb func(refreshingRequest *types.RefreshingRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.RefreshingRequestKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var refreshingRequest types.RefreshingRequest
		k.cdc.MustUnmarshal(iterator.Value(), &refreshingRequest)

		if cb(&refreshingRequest) {
			break
		}
	}
}

// SetRefreshingCompletion sets the given refreshing completion
func (k Keeper) SetRefreshingCompletion(ctx sdk.Context, completion *types.RefreshingCompletion) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(completion)
	store.Set(types.RefreshingCompletionKey(completion.Id, completion.ConsensusPubkey), bz)
}

// HasRefreshingCompletion returns true if the given completion exists, false otherwise
func (k Keeper) HasRefreshingCompletion(ctx sdk.Context, id uint64, consPubKey string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.RefreshingCompletionKey(id, consPubKey))
}

// GetRefreshingCompletions gets refreshing completions by the given id
func (k Keeper) GetRefreshingCompletions(ctx sdk.Context, id uint64) []*types.RefreshingCompletion {
	completions := make([]*types.RefreshingCompletion, 0)

	k.IterateRefreshingCompletions(ctx, id, func(completion *types.RefreshingCompletion) (stop bool) {
		completions = append(completions, completion)
		return false
	})

	return completions
}

// IterateRefreshingCompletions iterates through refreshing completions by the given id
func (k Keeper) IterateRefreshingCompletions(ctx sdk.Context, id uint64, cb func(completion *types.RefreshingCompletion) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.RefreshingCompletionKeyPrefix, sdk.Uint64ToBigEndian(id)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var req types.RefreshingCompletion
		k.cdc.MustUnmarshal(iterator.Value(), &req)

		if cb(&req) {
			break
		}
	}
}

// InitiateRefreshingRequest initiates the refreshing request with the specified params
func (k Keeper) InitiateRefreshingRequest(ctx sdk.Context, dkgId uint64, removedParticipants []string, timeoutDuration time.Duration) *types.RefreshingRequest {
	req := &types.RefreshingRequest{
		Id:                  k.IncrementRefreshingRequestId(ctx),
		DkgId:               dkgId,
		RemovedParticipants: removedParticipants,
		ExpirationTime:      types.GetExpirationTime(ctx.BlockTime(), timeoutDuration),
		Status:              types.RefreshingStatus_REFRESHING_STATUS_PENDING,
	}

	k.SetRefreshingRequest(ctx, req)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeInitiateRefreshing,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", req.Id)),
			sdk.NewAttribute(types.AttributeKeyDKGId, fmt.Sprintf("%d", dkgId)),
			sdk.NewAttribute(types.AttributeKeyRemovedParticipants, strings.Join(removedParticipants, types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AttributeKeyExpirationTime, req.ExpirationTime.String()),
		),
	)

	return req
}

// CompleteRefreshing completes the refreshing request by the participant
// The refreshing request will be finalized when all participants submit valid completions before timeout
func (k Keeper) CompleteRefreshing(ctx sdk.Context, sender string, id uint64, consensusPubKey string, signature string) error {
	if !k.HasRefreshingRequest(ctx, id) {
		return types.ErrRefreshingRequestDoesNotExist
	}

	refreshingRequest := k.GetRefreshingRequest(ctx, id)
	if refreshingRequest.Status != types.RefreshingStatus_REFRESHING_STATUS_PENDING {
		return errorsmod.Wrap(types.ErrInvalidRefreshingStatus, "refreshing request non pending")
	}

	if !refreshingRequest.ExpirationTime.IsZero() && !ctx.BlockTime().Before(refreshingRequest.ExpirationTime) {
		return types.ErrRefreshingRequestExpired
	}

	if !types.ParticipantExists(k.GetRefreshingParticipants(ctx, refreshingRequest), consensusPubKey) {
		return types.ErrUnauthorizedParticipant
	}

	if k.HasRefreshingCompletion(ctx, id, consensusPubKey) {
		return types.ErrRefreshingCompletionAlreadyExists
	}

	if !types.VerifySignature(signature, consensusPubKey, types.GetRefreshingCompletionSigMsg(id)) {
		return types.ErrInvalidSignature
	}

	completion := &types.RefreshingCompletion{
		Id:              id,
		Sender:          sender,
		ConsensusPubkey: consensusPubKey,
		Signature:       signature,
	}

	k.SetRefreshingCompletion(ctx, completion)

	return nil
}
