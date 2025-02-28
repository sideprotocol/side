package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// GetPrice gets the current price for the specified pair
func (k Keeper) GetPrice(ctx sdk.Context, pair string) sdkmath.Int {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.PriceKey(pair))
	price, ok := sdkmath.NewIntFromString(string(bz))
	if !ok {
		price = sdkmath.ZeroInt()
	}

	return price
}

// SetPrice sets the price for the specified pair
func (k Keeper) SetPrice(ctx sdk.Context, pair string, price string) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.PriceKey(pair), []byte(price))
}
