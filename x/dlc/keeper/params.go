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

// PriceIntervals gets all supported price intervals
func (k Keeper) PriceIntervals(ctx sdk.Context) []types.PriceInterval {
	return k.GetParams(ctx).PriceIntervals
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

// NonceGenerationBatchSize gets the nonce generation batch size
func (k Keeper) NonceGenerationBatchSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).NonceGenerationBatchSize
}

// DKGTimeoutPeriod gets the DKG timeout period
func (k Keeper) DKGTimeoutPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).DkgTimeoutPeriod
}
