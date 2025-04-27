package keeper

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/tss/types"
)

// GetResharingRequestId returns the resharing request id
func (k Keeper) GetResharingRequestId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.ResharingRequestIdKey)

	return sdk.BigEndianToUint64(bz)
}

// IncrementResharingRequestId increments the resharing request id and returns the new id
func (k Keeper) IncrementResharingRequestId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetResharingRequestId(ctx) + 1
	store.Set(types.ResharingRequestIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// SetResharingRequest sets the resharing request
func (k Keeper) SetResharingRequest(ctx sdk.Context, resharingRequest *types.ResharingRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(resharingRequest)

	store.Set(types.ResharingRequestKey(resharingRequest.Id), bz)
}

// HasResharingRequest returns true if the given resharing request exists, false otherwise
func (k Keeper) HasResharingRequest(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.ResharingRequestKey(id))
}

// GetResharingRequest returns the resharing request by the given id
func (k Keeper) GetResharingRequest(ctx sdk.Context, id uint64) *types.ResharingRequest {
	store := ctx.KVStore(k.storeKey)

	var resharingRequest types.ResharingRequest
	bz := store.Get(types.ResharingRequestKey(id))
	k.cdc.MustUnmarshal(bz, &resharingRequest)

	return &resharingRequest
}

// GetResharingRequests gets the resharing requests by the given status
func (k Keeper) GetResharingRequests(ctx sdk.Context, status types.ResharingStatus) []*types.ResharingRequest {
	requests := make([]*types.ResharingRequest, 0)

	k.IterateResharingRequests(ctx, func(req *types.ResharingRequest) (stop bool) {
		if req.Status == status {
			requests = append(requests, req)
		}

		return false
	})

	return requests
}

// GetPendingResharingRequests gets the pending resharing requests
func (k Keeper) GetPendingResharingRequests(ctx sdk.Context) []*types.ResharingRequest {
	requests := make([]*types.ResharingRequest, 0)

	k.IterateResharingRequests(ctx, func(req *types.ResharingRequest) (stop bool) {
		if req.Status == types.ResharingStatus_RESHARING_STATUS_PENDING {
			requests = append(requests, req)
		}

		return false
	})

	return requests
}

// IterateResharingRequests iterates through all resharing requests
func (k Keeper) IterateResharingRequests(ctx sdk.Context, cb func(resharingRequest *types.ResharingRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.ResharingRequestKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var resharingRequest types.ResharingRequest
		k.cdc.MustUnmarshal(iterator.Value(), &resharingRequest)

		if cb(&resharingRequest) {
			break
		}
	}
}

// SetResharingCompletion sets the given resharing completion
func (k Keeper) SetResharingCompletion(ctx sdk.Context, completion *types.ResharingCompletion) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(completion)
	store.Set(types.ResharingCompletionKey(completion.Id, completion.ConsensusPubkey), bz)
}

// HasResharingCompletion returns true if the given completion exists, false otherwise
func (k Keeper) HasResharingCompletion(ctx sdk.Context, id uint64, consPubKey string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.ResharingCompletionKey(id, consPubKey))
}

// GetResharingCompletions gets resharing completions by the given id
func (k Keeper) GetResharingCompletions(ctx sdk.Context, id uint64) []*types.ResharingCompletion {
	completions := make([]*types.ResharingCompletion, 0)

	k.IterateResharingCompletions(ctx, id, func(completion *types.ResharingCompletion) (stop bool) {
		completions = append(completions, completion)
		return false
	})

	return completions
}

// IterateResharingCompletions iterates through resharing completions by the given id
func (k Keeper) IterateResharingCompletions(ctx sdk.Context, id uint64, cb func(completion *types.ResharingCompletion) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.ResharingCompletionKeyPrefix, sdk.Uint64ToBigEndian(id)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var req types.ResharingCompletion
		k.cdc.MustUnmarshal(iterator.Value(), &req)

		if cb(&req) {
			break
		}
	}
}

// InitiateResharingRequest initiates the resharing request with the specified params
func (k Keeper) InitiateResharingRequest(ctx sdk.Context, dkgId uint64, pubKey string, participants []string) *types.ResharingRequest {
	req := &types.ResharingRequest{
		Id:             k.IncrementResharingRequestId(ctx),
		DkgId:          dkgId,
		PubKey:         pubKey,
		Participants:   participants,
		ExpirationTime: ctx.BlockTime().Add(k.DKGTimeoutPeriod(ctx)),
		Status:         types.ResharingStatus_RESHARING_STATUS_PENDING,
	}

	k.SetResharingRequest(ctx, req)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeInitiateResharing,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", req.Id)),
			sdk.NewAttribute(types.AttributeKeyDKGId, fmt.Sprintf("%d", dkgId)),
			sdk.NewAttribute(types.AttributeKeyPubKey, pubKey),
			sdk.NewAttribute(types.AttributeKeyParticipants, strings.Join(participants, types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AttributeKeyExpirationTime, req.ExpirationTime.String()),
		),
	)

	return req
}

// CompleteResharing completes the resharing request by the participant
// The resharing request will be finalized when all participants submit valid completions before timeout
func (k Keeper) CompleteResharing(ctx sdk.Context, sender string, id uint64, consensusPubKey string, signature string) error {
	if !k.HasResharingRequest(ctx, id) {
		return types.ErrResharingRequestDoesNotExist
	}

	resharingRequest := k.GetResharingRequest(ctx, id)
	if resharingRequest.Status != types.ResharingStatus_RESHARING_STATUS_PENDING {
		return errorsmod.Wrap(types.ErrInvalidResharingStatus, "resharing request non pending")
	}

	if !ctx.BlockTime().Before(resharingRequest.ExpirationTime) {
		return types.ErrResharingRequestExpired
	}

	if !types.ParticipantExists(resharingRequest.Participants, consensusPubKey) {
		return types.ErrUnauthorizedParticipant
	}

	if k.HasResharingCompletion(ctx, id, consensusPubKey) {
		return types.ErrResharingCompletionAlreadyExists
	}

	if !types.VerifySignature(signature, consensusPubKey, types.GetResharingCompletionSigMsg(id)) {
		return types.ErrInvalidSignature
	}

	completion := &types.ResharingCompletion{
		Id:              id,
		Sender:          sender,
		ConsensusPubkey: consensusPubKey,
		Signature:       signature,
	}

	k.SetResharingCompletion(ctx, completion)

	return nil
}
