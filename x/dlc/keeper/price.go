package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/dlc/types"
)

// GetPrice gets the current price for the specified pair
func (k Keeper) GetPrice(ctx sdk.Context, pair string) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.PriceKey(pair))
	return sdk.BigEndianToUint64(bz)
}

// SetPrice sets the price for the specified pair
func (k Keeper) SetPrice(ctx sdk.Context, pair string, price uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.PriceKey(pair), sdk.Uint64ToBigEndian(price))
}
