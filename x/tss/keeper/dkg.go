package keeper

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/tss/types"
)

// GetNextDKGRequestId gets the next DKG request ID
func (keeper Keeper) GetNextDKGRequestId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.DKGRequestIdKey)
	if bz == nil {
		return 1
	}

	return sdk.BigEndianToUint64(bz) + 1
}

// SetDKGRequestId sets the current DKG request ID
func (keeper Keeper) SetDKGRequestId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(keeper.storeKey)

	store.Set(types.DKGRequestIdKey, sdk.Uint64ToBigEndian(id))
}

// SetDKGRequest sets the given DKG request
func (k Keeper) SetDKGRequest(ctx sdk.Context, req *types.DKGRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(req)
	store.Set(types.DKGRequestKey(req.Id), bz)
}

// HasDKGRequest returns true if the given DKG exists, false otherwise
func (keeper Keeper) HasDKGRequest(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(keeper.storeKey)

	return store.Has(types.DKGRequestKey(id))
}

// GetDKGRequest gets the DKG request by the given id
func (k Keeper) GetDKGRequest(ctx sdk.Context, id uint64) *types.DKGRequest {
	store := ctx.KVStore(k.storeKey)

	var req types.DKGRequest
	bz := store.Get(types.DKGRequestKey(id))
	k.cdc.MustUnmarshal(bz, &req)

	return &req
}

// GetDKGRequests gets the DKG requests by the given status
func (k Keeper) GetDKGRequests(ctx sdk.Context, status types.DKGStatus) []*types.DKGRequest {
	requests := make([]*types.DKGRequest, 0)

	k.IterateDKGRequests(ctx, func(req *types.DKGRequest) (stop bool) {
		if req.Status == status {
			requests = append(requests, req)
		}

		return false
	})

	return requests
}

// GetPendingDKGRequests gets the pending DKG requests
func (k Keeper) GetPendingDKGRequests(ctx sdk.Context) []*types.DKGRequest {
	requests := make([]*types.DKGRequest, 0)

	k.IterateDKGRequests(ctx, func(req *types.DKGRequest) (stop bool) {
		if req.Status == types.DKGStatus_DKG_STATUS_PENDING {
			requests = append(requests, req)
		}

		return false
	})

	return requests
}

// GetAllDKGRequests gets all DKG requests
func (k Keeper) GetAllDKGRequests(ctx sdk.Context) []*types.DKGRequest {
	requests := make([]*types.DKGRequest, 0)

	k.IterateDKGRequests(ctx, func(req *types.DKGRequest) (stop bool) {
		requests = append(requests, req)
		return false
	})

	return requests
}

// IterateDKGRequests iterates through all DKG requests
func (k Keeper) IterateDKGRequests(ctx sdk.Context, cb func(req *types.DKGRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.DKGRequestKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var req types.DKGRequest
		k.cdc.MustUnmarshal(iterator.Value(), &req)

		if cb(&req) {
			break
		}
	}
}

// SetDKGCompletion sets the given DKG completion
func (k Keeper) SetDKGCompletion(ctx sdk.Context, completion *types.DKGCompletion) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(completion)
	store.Set(types.DKGCompletionKey(completion.Id, completion.ConsensusPubkey), bz)
}

// HasDKGCompletion returns true if the given completion exists, false otherwise
func (k Keeper) HasDKGCompletion(ctx sdk.Context, id uint64, consPubKey string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.DKGCompletionKey(id, consPubKey))
}

// GetDKGCompletions gets DKG completions by the given DKG request id
func (k Keeper) GetDKGCompletions(ctx sdk.Context, id uint64) []*types.DKGCompletion {
	completions := make([]*types.DKGCompletion, 0)

	k.IterateDKGCompletions(ctx, id, func(completion *types.DKGCompletion) (stop bool) {
		completions = append(completions, completion)
		return false
	})

	return completions
}

// IterateDKGCompletions iterates through DKG completions by the given DKG request id
func (k Keeper) IterateDKGCompletions(ctx sdk.Context, id uint64, cb func(completion *types.DKGCompletion) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.DKGCompletionKeyPrefix, sdk.Uint64ToBigEndian(id)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var req types.DKGCompletion
		k.cdc.MustUnmarshal(iterator.Value(), &req)

		if cb(&req) {
			break
		}
	}
}

// InitiateDKG initiates the DKG request by the specified params
func (k Keeper) InitiateDKG(ctx sdk.Context, module string, ty string, intent int32, participants []string, threshold uint32, batchSize uint32) *types.DKGRequest {
	req := &types.DKGRequest{
		Id:             k.GetNextDKGRequestId(ctx),
		Module:         module,
		Type:           ty,
		Intent:         intent,
		Participants:   participants,
		Threshold:      threshold,
		BatchSize:      batchSize,
		ExpirationTime: ctx.BlockTime().Add(k.DKGTimeoutPeriod(ctx)),
		Status:         types.DKGStatus_DKG_STATUS_PENDING,
	}

	k.SetDKGRequest(ctx, req)
	k.SetDKGRequestId(ctx, req.Id)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeInitiateDKG,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", req.Id)),
			sdk.NewAttribute(types.AttributeKeyModule, module),
			sdk.NewAttribute(types.AttributeKeyType, ty),
			sdk.NewAttribute(types.AttributeKeyIntent, fmt.Sprintf("%d", intent)),
			sdk.NewAttribute(types.AttributeKeyParticipants, strings.Join(participants, types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", threshold)),
			sdk.NewAttribute(types.AttributeKeyBatchSize, fmt.Sprintf("%d", batchSize)),
			sdk.NewAttribute(types.AttributeKeyExpirationTime, req.ExpirationTime.String()),
		),
	)

	return req
}

// CompleteDKG completes the DKG request by the DKG participant
// The DKG request will be completed when all participants submit valid completions before timeout
func (k Keeper) CompleteDKG(ctx sdk.Context, sender string, id uint64, pubKeys []string, consensusPubKey string, signature string) error {
	if !k.HasDKGRequest(ctx, id) {
		return types.ErrDKGRequestDoesNotExist
	}

	dkgRequest := k.GetDKGRequest(ctx, id)
	if dkgRequest.Status != types.DKGStatus_DKG_STATUS_PENDING {
		return errorsmod.Wrap(types.ErrInvalidDKGStatus, "dkg request non pending")
	}

	if !ctx.BlockTime().Before(dkgRequest.ExpirationTime) {
		return types.ErrDKGRequestExpired
	}

	if !types.ParticipantExists(dkgRequest.Participants, consensusPubKey) {
		return types.ErrUnauthorizedParticipant
	}

	if k.HasDKGCompletion(ctx, id, consensusPubKey) {
		return types.ErrDKGCompletionAlreadyExists
	}

	if len(pubKeys) != int(dkgRequest.BatchSize) {
		return errorsmod.Wrap(types.ErrInvalidDKGCompletion, "mismatched public key count")
	}

	if !types.VerifySignature(signature, consensusPubKey, types.GetSigMsg(id, pubKeys)) {
		return types.ErrInvalidSignature
	}

	completion := &types.DKGCompletion{
		Id:              id,
		Sender:          sender,
		PubKeys:         pubKeys,
		ConsensusPubkey: consensusPubKey,
		Signature:       signature,
	}

	k.SetDKGCompletion(ctx, completion)

	return nil
}
