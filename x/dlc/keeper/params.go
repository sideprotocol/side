package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/dlc/types"
)

// GetNonceQueueSize gets the nonce queue size
func (k Keeper) GetNonceQueueSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).NonceQueueSize
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

// GetDKGTimeoutPeriod gets the DKG timeout period
func (k Keeper) GetDKGTimeoutPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).DkgTimeoutPeriod
}
