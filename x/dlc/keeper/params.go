package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetPriceInterval gets the price interval for the given pair
func (k Keeper) GetPriceInterval(ctx sdk.Context, pair string) int32 {
	priceIntervals := k.GetParams(ctx).PriceIntervals

	for _, pi := range priceIntervals {
		if pi.PricePair == pair {
			return pi.Interval
		}
	}

	return 0
}

// GetDKGTimeoutPeriod gets the DKG timeout period
func (k Keeper) GetDKGTimeoutPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).DkgTimeoutPeriod
}
