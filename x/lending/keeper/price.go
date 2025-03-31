package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

func (k Keeper) SetPrice(ctx sdk.Context, price string) {
	store := ctx.KVStore(k.storeKey)

	if sdkmath.LegacyMustNewDecFromStr(price).IsZero() {
		store.Delete(types.PriceKey)
		return
	}

	store.Set(types.PriceKey, []byte(price))
}

func (k Keeper) GetPrice(ctx sdk.Context, pair string) (sdkmath.LegacyDec, error) {
	price, err := k.GetLocalPrice(ctx, pair)
	if err == nil {
		return price, nil
	}

	return k.oracleKeeper.GetPrice(ctx, pair)
}

func (k Keeper) GetLocalPrice(ctx sdk.Context, pair string) (sdkmath.LegacyDec, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.PriceKey)
	if bz == nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("no price set")
	}

	return sdkmath.LegacyNewDecFromStr(string(bz))
}
