package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
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

// IteratePrices iterates through all oracle prices
func (k Keeper) IteratePrices(ctx sdk.Context, process func(header types.OraclePrice) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PriceKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var header types.OraclePrice
		key := iterator.Key()
		header.Symbol = string(key[1:])
		bz := iterator.Value()
		price, _ := sdkmath.LegacyNewDecFromStr(string(bz))
		header.Price = price
		if process(header) {
			break
		}
	}
}

// GetAllPrices returns all oracle prices
func (k Keeper) GetAllPrices(ctx sdk.Context) []*types.OraclePrice {
	var prices []*types.OraclePrice
	k.IteratePrices(ctx, func(price types.OraclePrice) (stop bool) {
		prices = append(prices, &price)
		return false
	})
	return prices
}
