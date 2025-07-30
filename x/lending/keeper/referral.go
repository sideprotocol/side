package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/sideprotocol/side/x/lending/types"
)

// SetReferrer sets the given referrer
func (k Keeper) SetReferrer(ctx sdk.Context, referrer *types.Referrer) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(referrer)

	store.Set(types.ReferrerKey(referrer.ReferralCode), bz)
}

// HasReferrer returns true if the given referrer exists, false otherwise
func (k Keeper) HasReferrer(ctx sdk.Context, referralCode string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.ReferrerKey(referralCode))
}

// GetReferrer gets the given referrer
func (k Keeper) GetReferrer(ctx sdk.Context, referralCode string) *types.Referrer {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.ReferrerKey(referralCode))
	if bz == nil {
		return nil
	}

	var referrer types.Referrer
	k.cdc.MustUnmarshal(bz, &referrer)

	return &referrer
}

// GetReferrersWithPagination gets referrers with pagination
func (k Keeper) GetReferrersWithPagination(ctx sdk.Context, pagination *query.PageRequest) ([]*types.Referrer, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	referrerStore := prefix.NewStore(store, types.ReferrerKeyPrefix)

	var referrers []*types.Referrer

	pageRes, err := query.Paginate(referrerStore, pagination, func(key []byte, value []byte) error {
		var referrer types.Referrer
		k.cdc.MustUnmarshal(value, &referrer)

		referrers = append(referrers, &referrer)

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return referrers, pageRes, nil
}

// GetReferrers gets all referrers
func (k Keeper) GetAllReferrers(ctx sdk.Context) []*types.Referrer {
	var referrers []*types.Referrer

	k.IterateReferrers(ctx, func(referrer *types.Referrer) (stop bool) {
		referrers = append(referrers, referrer)
		return false
	})

	return referrers
}

// IterateReferrers iterates through all referrers
func (k Keeper) IterateReferrers(ctx sdk.Context, cb func(referrer *types.Referrer) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.ReferrerKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var referrer types.Referrer
		k.cdc.MustUnmarshal(iterator.Value(), &referrer)

		if cb(&referrer) {
			break
		}
	}
}
