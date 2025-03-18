package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/dlc/types"
)

// GetPriceEventNonceQueueSize gets the nonce queue size for the price events
func (k Keeper) GetPriceEventNonceQueueSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).PriceEventNonceQueueSize
}

// GetPriceInterval gets the price interval for the given pair
func (k Keeper) GetPriceInterval(ctx sdk.Context, pair string) int32 {
	priceIntervals := k.GetParams(ctx).PriceIntervals

	for _, pi := range priceIntervals {
		if pi.PricePair == pair {
			return pi.Interval
		}
	}

	return types.DefaultPriceInterval
}

// GetLendingEventInitialNonceCount gets the initial nonce count for the lending events
func (k Keeper) GetLendingEventInitialNonceCount(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).LendingEventInitialNonceCount
}

// GetLendingEventNonceUsageThreshold gets the nonce usage threshold for the lending events
func (k Keeper) GetLendingEventNonceUsageThreshold(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).LendingEventNonceUsageThreshold
}

// GetDKGTimeoutPeriod gets the DKG timeout period
func (k Keeper) GetDKGTimeoutPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).DkgTimeoutPeriod
}
