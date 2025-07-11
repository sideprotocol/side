package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/farming/types"
)

// GetPhaseId gets the current phase id
func (k Keeper) GetPhaseId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.PhaseIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementPhaseId increments the phase id
func (k Keeper) IncrementPhaseId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetPhaseId(ctx) + 1
	store.Set(types.PhaseIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// HasPhase returns true if the given phase exists, false otherwise
func (k Keeper) HasPhase(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.PhaseKey(id))
}

// GetPhase gets the phase by the given id
func (k Keeper) GetPhase(ctx sdk.Context, id uint64) *types.Phase {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.PhaseKey(id))
	var phase types.Phase
	k.cdc.MustUnmarshal(bz, &phase)

	return &phase
}

// SetPhase sets the given phase
func (k Keeper) SetPhase(ctx sdk.Context, phase *types.Phase) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(phase)

	store.Set(types.PhaseKey(phase.Id), bz)
}

// GetCurrentPhase gets the current phase
func (k Keeper) GetCurrentPhase(ctx sdk.Context) *types.Phase {
	id := k.GetPhaseId(ctx)
	if id == 0 {
		return nil
	}

	return k.GetPhase(ctx, id)
}

// GetAllPhases gets all phases
func (k Keeper) GetAllPhases(ctx sdk.Context) []*types.Phase {
	phases := make([]*types.Phase, 0)

	k.IteratePhases(ctx, func(phase *types.Phase) (stop bool) {
		phases = append(phases, phase)
		return false
	})

	return phases
}

// IteratePhases iterates through all phases
func (k Keeper) IteratePhases(ctx sdk.Context, cb func(phase *types.Phase) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.PhaseKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var phase types.Phase
		k.cdc.MustUnmarshal(iterator.Value(), &phase)

		if cb(&phase) {
			break
		}
	}
}
