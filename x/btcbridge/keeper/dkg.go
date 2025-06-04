package keeper

import (
	"slices"
	"time"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// GetNextDKGRequestID gets the next DKG request ID
func (keeper Keeper) GetNextDKGRequestID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.DKGRequestIDKey)
	if bz == nil {
		return 1
	}

	return sdk.BigEndianToUint64(bz) + 1
}

// SetDKGRequestID sets the current DKG request ID
func (keeper Keeper) SetDKGRequestID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(keeper.storeKey)

	store.Set(types.DKGRequestIDKey, sdk.Uint64ToBigEndian(id))
}

// SetDKGRequest sets the given DKG request
func (k Keeper) SetDKGRequest(ctx sdk.Context, req *types.DKGRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(req)
	store.Set(types.DKGRequestKey(req.Id), bz)
}

// HasDKGRequest returns true if the given DKG request exists, false otherwise
func (k Keeper) HasDKGRequest(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

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
func (k Keeper) GetDKGRequests(ctx sdk.Context, status types.DKGRequestStatus) []*types.DKGRequest {
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
		if req.Status == types.DKGRequestStatus_DKG_REQUEST_STATUS_PENDING {
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

// GetDKGRequestExpirationTime gets the expiration time of the DKG request
func (k Keeper) GetDKGRequestExpirationTime(ctx sdk.Context) *time.Time {
	creationTime := ctx.BlockTime()
	timeout := k.GetParams(ctx).TssParams.DkgTimeoutPeriod

	expiration := creationTime.Add(timeout)

	return &expiration
}

// SetDKGCompletionRequest sets the given DKG completion request
func (k Keeper) SetDKGCompletionRequest(ctx sdk.Context, req *types.DKGCompletionRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(req)
	store.Set(types.DKGCompletionRequestKey(req.Id, req.ConsensusPubkey), bz)
}

// HasDKGCompletionRequest returns true if the given completion request exists, false otherwise
func (k Keeper) HasDKGCompletionRequest(ctx sdk.Context, id uint64, consPubKey string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.DKGCompletionRequestKey(id, consPubKey))
}

// GetDKGCompletionRequests gets DKG completion requests by the given id
func (k Keeper) GetDKGCompletionRequests(ctx sdk.Context, id uint64) []*types.DKGCompletionRequest {
	requests := make([]*types.DKGCompletionRequest, 0)

	k.IterateDKGCompletionRequests(ctx, id, func(req *types.DKGCompletionRequest) (stop bool) {
		requests = append(requests, req)
		return false
	})

	return requests
}

// IterateDKGCompletionRequests iterates through all DKG completion requests by the given id
func (k Keeper) IterateDKGCompletionRequests(ctx sdk.Context, id uint64, cb func(req *types.DKGCompletionRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.DKGCompletionRequestKeyPrefix, sdk.Uint64ToBigEndian(id)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var req types.DKGCompletionRequest
		k.cdc.MustUnmarshal(iterator.Value(), &req)

		if cb(&req) {
			break
		}
	}
}

// InitiateDKG initiates the DKG request by the specified params
func (k Keeper) InitiateDKG(ctx sdk.Context, participants []*types.DKGParticipant, threshold uint32, vaultTypes []types.AssetType, enableTransfer bool, targetUtxoNum uint32) (*types.DKGRequest, error) {
	baseParticipants := k.tssKeeper.AllowedDKGParticipants(ctx)

	if len(baseParticipants) != 0 {
		for _, p := range participants {
			if !slices.Contains(baseParticipants, p.ConsensusPubkey) {
				return nil, errorsmod.Wrap(types.ErrInvalidDKGParams, "participant not authorized")
			}
		}
	}

	req := &types.DKGRequest{
		Id:             k.GetNextDKGRequestID(ctx),
		Participants:   participants,
		Threshold:      threshold,
		VaultTypes:     vaultTypes,
		EnableTransfer: enableTransfer,
		TargetUtxoNum:  targetUtxoNum,
		Expiration:     k.GetDKGRequestExpirationTime(ctx),
		Status:         types.DKGRequestStatus_DKG_REQUEST_STATUS_PENDING,
	}

	k.SetDKGRequest(ctx, req)
	k.SetDKGRequestID(ctx, req.Id)

	return req, nil
}

// CompleteDKG completes the DKG request by the DKG participant
// The DKG request will be finalized when all participants submit the valid completion request before timeout
func (k Keeper) CompleteDKG(ctx sdk.Context, req *types.DKGCompletionRequest) error {
	if !k.HasDKGRequest(ctx, req.Id) {
		return types.ErrDKGRequestDoesNotExist
	}

	dkgReq := k.GetDKGRequest(ctx, req.Id)
	if !types.ParticipantExists(dkgReq.Participants, req.ConsensusPubkey) {
		return types.ErrUnauthorizedDKGCompletionRequest
	}

	if k.HasDKGCompletionRequest(ctx, req.Id, req.ConsensusPubkey) {
		return types.ErrDKGCompletionRequestExists
	}

	if dkgReq.Status != types.DKGRequestStatus_DKG_REQUEST_STATUS_PENDING {
		return errorsmod.Wrap(types.ErrInvalidDKGCompletionRequest, "invalid dkg request status")
	}

	if !ctx.BlockTime().Before(*dkgReq.Expiration) {
		return errorsmod.Wrap(types.ErrInvalidDKGCompletionRequest, "dkg request expired")
	}

	if err := k.CheckVaults(ctx, req.Vaults, dkgReq.VaultTypes); err != nil {
		return err
	}

	if !types.VerifySignature(req.Signature, req.ConsensusPubkey, types.GetDKGCompletionSigMsg(req)) {
		return errorsmod.Wrap(types.ErrInvalidDKGCompletionRequest, "invalid signature")
	}

	k.SetDKGCompletionRequest(ctx, req)

	return nil
}
