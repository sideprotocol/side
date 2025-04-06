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
