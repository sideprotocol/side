package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetPrice gets the current price of the given pair
func (k Keeper) GetPrice(ctx sdk.Context, pair string) (sdkmath.LegacyDec, error) {
	return k.oracleKeeper.GetPrice(ctx, pair)
}
