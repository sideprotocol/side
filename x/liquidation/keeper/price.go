package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/types"
)

// GetPrice gets the current price for the specified pair
func (k Keeper) GetPrice(ctx sdk.Context, pair string) sdkmath.LegacyDec {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.PriceKey(pair))
	price, err := sdkmath.LegacyNewDecFromStr(string(bz))
	if err != nil {
		price = sdkmath.LegacyZeroDec()
	}

	return price
}

// SetPrice sets the price for the specified pair
func (k Keeper) SetPrice(ctx sdk.Context, pair string, price string) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.PriceKey(pair), []byte(price))
}
