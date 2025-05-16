package dlc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/keeper"
	"github.com/sideprotocol/side/x/dlc/types"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	// check if there exist oracle participant base set
	if len(k.OracleParticipantBaseSet(ctx)) == 0 {
		return
	}

	generatePriceEventNonces(ctx, k)
	generateDateEventNonces(ctx, k)
	generateLendingEventNonces(ctx, k)
}

// generatePriceEventNonces generates nonces for dlc price events
func generatePriceEventNonces(ctx sdk.Context, k keeper.Keeper) {
	// check all supported price pairs
	for i, pi := range k.PriceIntervals(ctx) {
		// get current price
		currentPrice, err := k.GetPrice(ctx, pi.PricePair)
		if err != nil {
			k.Logger(ctx).Info("failed to get price", "pair", pi.PricePair, "err", err)
			continue
		}

		nonceQueueSize := int64(k.PriceEventNonceQueueSize(ctx))

		// check if price event nonces need to be generated
		currentEventPrice := k.GetCurrentEventPrice(ctx, pi.PricePair)
		if currentEventPrice.GTE(currentPrice.Add(pi.Interval.MulInt64(nonceQueueSize))) && k.GetTriggeredPriceEventQueueCount(ctx, pi.PricePair) == 0 {
			continue
		}

		// immediate generation required if there exist pending triggered price events
		if k.GetTriggeredPriceEventQueueCount(ctx, pi.PricePair) == 0 {
			// check block height
			if ctx.BlockHeight()%k.NonceGenerationInterval(ctx) != 0 {
				continue
			}
		}

		// initiate DKG
		k.TSSKeeper().InitiateDKG(ctx, types.ModuleName, types.DKG_TYPE_NONCE, int32(types.DKGIntent_DKG_INTENT_PRICE_EVENT_NONCE)+int32(i), k.GetOracleParticipants(ctx), k.OracleParticipantThreshold(ctx), k.NonceGenerationBatchSize(ctx))
	}
}

// generateDateEventNonces generates nonces for dlc date events
func generateDateEventNonces(ctx sdk.Context, k keeper.Keeper) {
	// check block height
	if ctx.BlockHeight()%k.NonceGenerationInterval(ctx) != 0 {
		return
	}

	// check if date event nonces need to be generated
	currentEventDate := k.GetCurrentEventDate(ctx)
	if (currentEventDate-ctx.BlockTime().Unix())/k.DateInterval(ctx) >= int64(k.DateEventNonceQueueSize(ctx)) {
		return
	}

	// initiate DKG
	k.TSSKeeper().InitiateDKG(ctx, types.ModuleName, types.DKG_TYPE_NONCE, int32(types.DKGIntent_DKG_INTENT_DATE_EVENT_NONCE), k.GetOracleParticipants(ctx), k.OracleParticipantThreshold(ctx), k.NonceGenerationBatchSize(ctx))
}

// generateLendingEventNonces generates nonces events for dlc lending events
func generateLendingEventNonces(ctx sdk.Context, k keeper.Keeper) {
	// check block height
	if ctx.BlockHeight()%k.NonceGenerationInterval(ctx) != 0 {
		return
	}

	// check if lending event nonces need to be generated
	pendingLendingEventCount := k.GetPendingLendingEventCount(ctx)
	if pendingLendingEventCount >= k.LendingEventNonceQueueSize(ctx) {
		return
	}

	// initiate DKG
	k.TSSKeeper().InitiateDKG(ctx, types.ModuleName, types.DKG_TYPE_NONCE, int32(types.DKGIntent_DKG_INTENT_LENDING_EVENT_NONCE), k.GetOracleParticipants(ctx), k.OracleParticipantThreshold(ctx), k.NonceGenerationBatchSize(ctx))
}
