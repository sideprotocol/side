package keeper

import (
	"fmt"
	"slices"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

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

	k.SetDKGRequestStatus(ctx, req.Id, req.Status)

	store.Set(types.DKGRequestKey(req.Id), bz)
}

// SetDKGRequestStatus sets the status store of the given DKG request
func (k Keeper) SetDKGRequestStatus(ctx sdk.Context, id uint64, status types.DKGStatus) {
	store := ctx.KVStore(k.storeKey)

	if k.HasDKGRequest(ctx, id) {
		k.RemoveDKGRequestStatus(ctx, id)
	}

	store.Set(types.DKGRequestByStatusKey(status, id), []byte{})
}

// RemoveDKGRequestStatus removes the status store of the given DKG request
func (k Keeper) RemoveDKGRequestStatus(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)

	dkgRequest := k.GetDKGRequest(ctx, id)

	store.Delete(types.DKGRequestByStatusKey(dkgRequest.Status, id))
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

// GetDKGRequestsByStatus gets the DKG requests by the given status
func (k Keeper) GetDKGRequestsByStatus(ctx sdk.Context, status types.DKGStatus) []*types.DKGRequest {
	requests := make([]*types.DKGRequest, 0)

	k.IterateDKGRequestsByStatus(ctx, status, func(req *types.DKGRequest) (stop bool) {
		requests = append(requests, req)
		return false
	})

	return requests
}

// GetDKGRequestsByStatusWithPagination gets the DKG requests by the given status and module with pagination
func (k Keeper) GetDKGRequestsByStatusWithPagination(ctx sdk.Context, status types.DKGStatus, module string, pagination *query.PageRequest) ([]*types.DKGRequest, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	dkgRequestStatusStore := prefix.NewStore(store, append(types.DKGRequestByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...))

	var dkgRequests []*types.DKGRequest

	pageRes, err := query.Paginate(dkgRequestStatusStore, pagination, func(key []byte, value []byte) error {
		id := sdk.BigEndianToUint64(key)
		dkgRequest := k.GetDKGRequest(ctx, id)

		if len(module) == 0 || dkgRequest.Module == module {
			dkgRequests = append(dkgRequests, dkgRequest)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return dkgRequests, pageRes, nil
}

// GetDKGRequestsWithPagination gets the DKG requests by the given module with pagination
func (k Keeper) GetDKGRequestsWithPagination(ctx sdk.Context, module string, pagination *query.PageRequest) ([]*types.DKGRequest, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	dkgRequestStore := prefix.NewStore(store, types.DKGRequestKeyPrefix)

	var dkgRequests []*types.DKGRequest

	pageRes, err := query.Paginate(dkgRequestStore, pagination, func(key []byte, value []byte) error {
		var dkgRequest types.DKGRequest
		k.cdc.MustUnmarshal(value, &dkgRequest)

		if len(module) == 0 || dkgRequest.Module == module {
			dkgRequests = append(dkgRequests, &dkgRequest)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return dkgRequests, pageRes, nil
}

// GetPendingDKGRequests gets the pending DKG requests
func (k Keeper) GetPendingDKGRequests(ctx sdk.Context) []*types.DKGRequest {
	return k.GetDKGRequestsByStatus(ctx, types.DKGStatus_DKG_STATUS_PENDING)
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

// IterateDKGRequestsByStatus iterates through DKG requests by the given status
func (k Keeper) IterateDKGRequestsByStatus(ctx sdk.Context, status types.DKGStatus, cb func(req *types.DKGRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	keyPrefix := append(types.DKGRequestByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...)

	iterator := storetypes.KVStorePrefixIterator(store, keyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		id := sdk.BigEndianToUint64(key[len(keyPrefix):])
		dkgRequest := k.GetDKGRequest(ctx, id)

		if cb(dkgRequest) {
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

// GetCompletedDKGParticipants gets participants that completed the given DKG request
func (k Keeper) GetCompletedDKGParticipants(ctx sdk.Context, id uint64) []string {
	participants := []string{}

	completions := k.GetDKGCompletions(ctx, id)
	for _, completion := range completions {
		participants = append(participants, completion.ConsensusPubkey)
	}

	return participants
}

// GetAbsentDKGParticipants gets the absent participants for the given DKG request
func (k Keeper) GetAbsentDKGParticipants(ctx sdk.Context, req *types.DKGRequest) []string {
	completedParticipants := k.GetCompletedDKGParticipants(ctx, req.Id)
	if len(completedParticipants) == 0 {
		return req.Participants
	}

	absentParticipants := []string{}
	for _, participant := range req.Participants {
		if !slices.Contains(completedParticipants, participant) {
			absentParticipants = append(absentParticipants, participant)
		}
	}

	return absentParticipants
}

// GetDKGPubKeys gets the pub keys generated by the given DKG
// Assume that the given DKG is completed
func (k Keeper) GetDKGPubKeys(ctx sdk.Context, id uint64) []string {
	return k.GetDKGCompletions(ctx, id)[0].PubKeys
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
		ExpirationTime: types.GetExpirationTime(ctx.BlockTime(), k.DKGTimeoutDuration(ctx)),
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

	if !types.VerifySignature(signature, consensusPubKey, types.GetDKGCompletionSigMsg(id, pubKeys)) {
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
