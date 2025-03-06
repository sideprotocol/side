package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/oracle/types"
)

func (k Keeper) HasPrice(ctx sdk.Context, symbol string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.PriceKey(symbol))
}

func (k Keeper) SetPrice(ctx sdk.Context, symbol, price string) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.PriceKey(symbol), []byte(price))
}

func (k Keeper) GetPrice(ctx sdk.Context, symbol string) (sdkmath.LegacyDec, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.PriceKey(symbol))
	if bz == nil {
		return sdkmath.LegacyZeroDec(), fmt.Errorf("no price set")
	}

	price, _ := sdkmath.LegacyNewDecFromStr(string(bz))
	return price, nil
}
