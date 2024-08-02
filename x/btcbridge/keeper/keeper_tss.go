package keeper

import (
	"time"

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

	iterator := sdk.KVStorePrefixIterator(store, types.DKGRequestKeyPrefix)
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

	expiration := creationTime.Add(*timeout)

	return &expiration
}

// SetDKGCompletionRequest sets the given DKG completion request
func (k Keeper) SetDKGCompletionRequest(ctx sdk.Context, req *types.DKGCompletionRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(req)
	store.Set(types.DKGCompletionRequestKey(req.Id, req.Validator), bz)
}

// HasDKGCompletionRequest returns true if the given completion request exists, false otherwise
func (k Keeper) HasDKGCompletionRequest(ctx sdk.Context, id uint64, validator string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.DKGCompletionRequestKey(id, validator))
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

	iterator := sdk.KVStorePrefixIterator(store, append(types.DKGCompletionRequestKeyPrefix, types.Int64ToBytes(id)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var req types.DKGCompletionRequest
		k.cdc.MustUnmarshal(iterator.Value(), &req)

		if cb(&req) {
			break
		}
	}
}

// CompleteDKG attempts to complete the DKG request
// The DKG request is completed when all participants submit the valid completion request before timeout
func (k Keeper) CompleteDKG(ctx sdk.Context, req *types.DKGCompletionRequest) error {
	dkgReq := k.GetDKGRequest(ctx, req.Id)
	if dkgReq == nil {
		return types.ErrDKGRequestDoesNotExist
	}

	if !types.ParticipantExists(dkgReq.Participants, req.Validator) {
		return types.ErrUnauthorizedDKGCompletionRequest
	}

	if k.HasDKGCompletionRequest(ctx, req.Id, req.Validator) {
		return types.ErrDKGCompletionRequestExists
	}

	if dkgReq.Status != types.DKGRequestStatus_DKG_REQUEST_STATUS_PENDING {
		return types.ErrInvalidDKGCompletionRequest
	}

	if !ctx.BlockTime().Before(*dkgReq.Expiration) {
		return types.ErrInvalidDKGCompletionRequest
	}

	if err := k.CheckVaults(ctx, req.Vaults); err != nil {
		return err
	}

	validatorConsAddr, _ := sdk.ConsAddressFromHex(req.Validator)
	validator, found := k.stakingKeeper.GetValidatorByConsAddr(ctx, validatorConsAddr)
	if !found {
		return types.ErrInvalidDKGCompletionRequest
	}

	pubKey, err := validator.ConsPubKey()
	if err != nil {
		return err
	}

	if !types.VerifySignature(req.Signature, pubKey.Bytes(), req) {
		return types.ErrInvalidDKGCompletionRequest
	}

	k.SetDKGCompletionRequest(ctx, req)

	return nil
}

// CheckVaults checks if the provided vaults are valid
func (k Keeper) CheckVaults(ctx sdk.Context, vaults []string) error {
	currentVaults := k.GetParams(ctx).Vaults

	if len(vaults) != len(currentVaults) {
		return types.ErrInvalidDKGCompletionRequest
	}

	for _, v := range vaults {
		if types.SelectVaultByBitcoinAddress(currentVaults, v) != nil {
			return types.ErrInvalidDKGCompletionRequest
		}
	}

	return nil
}

// UpdateVaults updates the asset vaults of the btc bridge
func (k Keeper) UpdateVaults(ctx sdk.Context, newVaults []string) {
	params := k.GetParams(ctx)

	for i, v := range newVaults {
		newVault := &types.Vault{
			Address:   v,
			AssetType: params.Vaults[i].AssetType,
			Version:   k.IncreaseVaultVersion(ctx),
		}

		params.Vaults = append(params.Vaults, newVault)
	}

	k.SetParams(ctx, params)
}

// IncreaseVaultVersion increases the vault version by 1
func (k Keeper) IncreaseVaultVersion(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	version := uint64(0)

	bz := store.Get(types.VaultVersionKey)
	if bz != nil {
		version = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.VaultVersionKey, sdk.Uint64ToBigEndian(version+1))

	return version + 1
}
