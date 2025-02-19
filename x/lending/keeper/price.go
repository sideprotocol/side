package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

func (k Keeper) SetPrice(ctx sdk.Context, price string) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.PriceKey, []byte(price))
}

func (k Keeper) GetPrice(ctx sdk.Context, pair string) (sdkmath.Int, error) {
	if k.oracleKeeper == nil {
		return k.GetLocalPrice(ctx, pair)
	}

	return k.oracleKeeper.GetPrice(ctx, pair)
}

func (k Keeper) GetLocalPrice(ctx sdk.Context, pair string) (sdkmath.Int, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.PriceKey)
	if bz == nil {
		return sdkmath.Int{}, fmt.Errorf("no price set")
	}

	price, _ := sdkmath.NewIntFromString(string(bz))
	return price, nil
}
