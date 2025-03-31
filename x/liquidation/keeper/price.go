package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/types"
)

// GetPrice gets the current price for the specified pair
func (k Keeper) GetPrice(ctx sdk.Context, pair string) (sdkmath.LegacyDec, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.PriceKey(pair))
	if bz == nil {
		return k.oracleKeeper.GetPrice(ctx, pair)
	}

	return sdkmath.LegacyNewDecFromStr(string(bz))
}

// SetPrice sets the price for the specified pair
func (k Keeper) SetPrice(ctx sdk.Context, pair string, price string) {
	store := ctx.KVStore(k.storeKey)

	if sdkmath.LegacyMustNewDecFromStr(price).IsZero() {
		store.Delete(types.PriceKey(pair))
		return
	}

	store.Set(types.PriceKey(pair), []byte(price))
}
