package dlc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/keeper"
	"github.com/sideprotocol/side/x/dlc/types"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
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
		if currentEventPrice >= currentPrice.TruncateInt64()+nonceQueueSize*int64(pi.Interval) && k.GetTriggeredPriceEventQueueCount(ctx) == 0 {
			continue
		}

		// get participants
		participants, err := k.GetOracleParticipants(ctx)
		if err != nil {
			k.Logger(ctx).Info("failed to get oracle participants", "err", err)
			return
		}

		threshold := len(participants) * 2 / 3

		// initiate DKG
		k.TSSKeeper().InitiateDKG(ctx, types.ModuleName, types.DKG_TYPE_NONCE, int32(types.DKGIntent_DKG_INTENT_PRICE_EVENT_NONCE)+int32(i), participants, uint32(threshold), k.NonceGenerationBatchSize(ctx))
	}
}

// generateDateEventNonces generates nonces for dlc date events
func generateDateEventNonces(ctx sdk.Context, k keeper.Keeper) {
	// check if date event nonces need to be generated
	currentEventDate := k.GetCurrentEventDate(ctx)
	if (currentEventDate-ctx.BlockTime().Unix())/k.DateInterval(ctx) >= int64(k.DateEventNonceQueueSize(ctx)) {
		return
	}

	// get participants
	participants, err := k.GetOracleParticipants(ctx)
	if err != nil {
		k.Logger(ctx).Info("failed to get oracle participants", "err", err)
		return
	}

	threshold := len(participants) * 2 / 3

	// initiate DKG
	k.TSSKeeper().InitiateDKG(ctx, types.ModuleName, types.DKG_TYPE_NONCE, int32(types.DKGIntent_DKG_INTENT_DATE_EVENT_NONCE), participants, uint32(threshold), k.NonceGenerationBatchSize(ctx))
}

// generateLendingEventNonces generates nonces events for dlc lending events
func generateLendingEventNonces(ctx sdk.Context, k keeper.Keeper) {
	// check if lending event nonces need to be generated
	pendingLendingEventCount := k.GetPendingLendingEventCount(ctx)
	if pendingLendingEventCount >= k.LendingEventNonceQueueSize(ctx) {
		return
	}

	// get participants
	participants, err := k.GetOracleParticipants(ctx)
	if err != nil {
		k.Logger(ctx).Info("failed to get oracle participants", "err", err)
		return
	}

	threshold := len(participants) * 2 / 3

	// initiate DKG
	k.TSSKeeper().InitiateDKG(ctx, types.ModuleName, types.DKG_TYPE_NONCE, int32(types.DKGIntent_DKG_INTENT_LENDING_EVENT_NONCE), participants, uint32(threshold), k.NonceGenerationBatchSize(ctx))
}
