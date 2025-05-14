package keeper

import (
	"time"

	sdkmath "cosmossdk.io/math"
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

// PriceInterval gets the price interval by the given pair
func (k Keeper) PriceInterval(ctx sdk.Context, pair string) sdkmath.LegacyDec {
	for _, pi := range k.PriceIntervals(ctx) {
		if pi.PricePair == pair {
			return pi.Interval
		}
	}

	return sdkmath.LegacyOneDec()
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

// OracleParticipantBaseSet gets the oracle participant base set
func (k Keeper) OracleParticipantBaseSet(ctx sdk.Context) []string {
	if len(k.GetParams(ctx).AllowedOracleParticipants) != 0 {
		return k.GetParams(ctx).AllowedOracleParticipants
	}

	return k.tssKeeper.GetParams(ctx).AllowedDkgParticipants
}

// OracleParticipantNum gets the oracle participant number
func (k Keeper) OracleParticipantNum(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).OracleParticipantNum
}

// OracleParticipantThreshold gets the oracle participant threshold
func (k Keeper) OracleParticipantThreshold(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).OracleParticipantThreshold
}

// NonceGenerationBatchSize gets the nonce generation batch size
func (k Keeper) NonceGenerationBatchSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).NonceGenerationBatchSize
}
