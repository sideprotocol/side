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

// GetDateEventNonceQueueSize gets the nonce queue size for the date events
func (k Keeper) GetDateEventNonceQueueSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).DateEventNonceQueueSize
}

// GetDateInterval gets the date interval for the date events in seconds
func (k Keeper) GetDateInterval(ctx sdk.Context) int64 {
	return int64(k.GetParams(ctx).DateInterval / time.Second)
}

// GetLendingEventNonceQueueSize gets the nonce queue size for the lending events
func (k Keeper) GetLendingEventNonceQueueSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).LendingEventNonceQueueSize
}

// GetDKGTimeoutPeriod gets the DKG timeout period
func (k Keeper) GetDKGTimeoutPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).DkgTimeoutPeriod
}
