package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// HandleConsolidation handles the vault consolidation request
func (k Keeper) HandleConsolidation(ctx sdk.Context, consolidation *types.Consolidation) error {
	if !types.HasVaultVersion(k.GetParams(ctx).Vaults, consolidation.VaultVersion) {
		return types.ErrInvalidConsolidation
	}

	k.SetConsolidation(ctx, consolidation)

	return nil
}

// IncreaseConsolidationID increases the consolidation id by 1
func (k Keeper) IncreaseConsolidationID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := uint64(0)

	bz := store.Get(types.ConsolidationIDKey)
	if bz != nil {
		id = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.VaultVersionKey, sdk.Uint64ToBigEndian(id+1))

	return id + 1
}

// SetConsolidation sets the given consolidation
func (k Keeper) SetConsolidation(ctx sdk.Context, consolidation *types.Consolidation) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(consolidation)
	store.Set(types.ConsolidationKey(consolidation.Id), bz)
}

// GetConsolidations gets all the consolidations
func (k Keeper) GetConsolidations(ctx sdk.Context) []*types.Consolidation {
	consolidations := make([]*types.Consolidation, 0)

	k.IterateConsolidations(ctx, func(c *types.Consolidation) (stop bool) {
		consolidations = append(consolidations, c)
		return false
	})

	return consolidations
}

// IterateConsolidations iterates over all the consolidations
func (k Keeper) IterateConsolidations(ctx sdk.Context, cb func(c *types.Consolidation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.ConsolidationKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var c types.Consolidation
		k.cdc.MustUnmarshal(iterator.Value(), &c)

		if cb(&c) {
			break
		}
	}
}
