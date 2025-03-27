package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// PriceEventNonceQueueSize gets the nonce queue size for the price events
func (k Keeper) PriceEventNonceQueueSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).PriceEventNonceQueueSize
}

// PriceInterval gets the price interval for the given pair
func (k Keeper) PriceInterval(ctx sdk.Context, pair string) int32 {
	priceIntervals := k.GetParams(ctx).PriceIntervals

	for _, pi := range priceIntervals {
		if pi.PricePair == pair {
			return pi.Interval
		}
	}

	return types.DefaultPriceInterval
}

// DateEventNonceQueueSize gets the nonce queue size for the date events
func (k Keeper) DateEventNonceQueueSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).DateEventNonceQueueSize
}

// DateInterval gets the date interval for the date events in seconds
func (k Keeper) DateInterval(ctx sdk.Context) int64 {
	return int64(k.GetParams(ctx).DateInterval / time.Second)
}

// LendingEventNonceQueueSize gets the nonce queue size for the lending events
func (k Keeper) LendingEventNonceQueueSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).LendingEventNonceQueueSize
}

// DKGTimeoutPeriod gets the DKG timeout period
func (k Keeper) DKGTimeoutPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).DkgTimeoutPeriod
}
