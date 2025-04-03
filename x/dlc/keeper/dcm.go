package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// CreateDCM creates a new DCM with the given pub key
func (k Keeper) CreateDCM(ctx sdk.Context, pubKey string) {
	dcm := &types.DCM{
		Id:     k.IncrementDCMId(ctx),
		Pubkey: pubKey,
		Time:   ctx.BlockTime(),
		Status: types.DCMStatus_DCM_status_Enable,
	}

	k.SetDCM(ctx, dcm)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateDCM,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", dcm.Id)),
			sdk.NewAttribute(types.AttributeKeyPubKey, dcm.Pubkey),
		),
	)
}

// GetDCMId gets the current DCM id
func (k Keeper) GetDCMId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.DCMIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementDCMId increments the DCM id and returns the new id
func (k Keeper) IncrementDCMId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetDCMId(ctx) + 1
	store.Set(types.DCMIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// HasDCM returns true if the given DCM exists, false otherwise
func (k Keeper) HasDCM(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.DCMKey(id))
}

// GetDCM gets the DCM by the given id
func (k Keeper) GetDCM(ctx sdk.Context, id uint64) *types.DCM {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.DCMKey(id))
	var dcm types.DCM
	k.cdc.MustUnmarshal(bz, &dcm)

	return &dcm
}

// SetDCM sets the given DCM
func (k Keeper) SetDCM(ctx sdk.Context, dcm *types.DCM) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(dcm)
	store.Set(types.DCMKey(dcm.Id), bz)
}

// HasPendingDCMPubKey returns true if the given pending DCM pubkey exists, false otherwise
func (k Keeper) HasPendingDCMPubKey(ctx sdk.Context, dcmId uint64, pubKey []byte) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.PendingDCMPubKeyKey(dcmId, pubKey))
}

// SetPendingDCMPubKey sets the pending DCM public key
func (k Keeper) SetPendingDCMPubKey(ctx sdk.Context, dcmId uint64, pubKey []byte, dcmPubKey []byte) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.PendingDCMPubKeyKey(dcmId, pubKey), dcmPubKey)
}

// GetDCMs gets DCMs by the given status
func (k Keeper) GetDCMs(ctx sdk.Context, status types.DCMStatus) []*types.DCM {
	dcms := make([]*types.DCM, 0)

	k.IterateDCMs(ctx, func(dcm *types.DCM) (stop bool) {
		if dcm.Status == status {
			dcms = append(dcms, dcm)
		}

		return false
	})

	return dcms
}

// GetPendingDCMPubKeys gets pending DCM pub keys by the given DCM id
func (k Keeper) GetPendingDCMPubKeys(ctx sdk.Context, dcmId uint64) [][]byte {
	pubKeys := make([][]byte, 0)

	k.IteratePendingDCMPubKeys(ctx, dcmId, func(pubKey []byte) (stop bool) {
		pubKeys = append(pubKeys, pubKey)

		return false
	})

	return pubKeys
}

// IterateDCMs iterates through all DCMs
func (k Keeper) IterateDCMs(ctx sdk.Context, cb func(dcm *types.DCM) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.DCMKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var dcm types.DCM
		k.cdc.MustUnmarshal(iterator.Value(), &dcm)

		if cb(&dcm) {
			break
		}
	}
}

// IteratePendingDCMPubKeys iterates through all pending DCM pub keys by the given DCM id
func (k Keeper) IteratePendingDCMPubKeys(ctx sdk.Context, dcmId uint64, cb func(pubKey []byte) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.PendingDCMPubKeyKeyPrefix, sdk.Uint64ToBigEndian(dcmId)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		if cb(iterator.Value()) {
			break
		}
	}
}
