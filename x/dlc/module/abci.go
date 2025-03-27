package dlc

import (
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/keeper"
	"github.com/sideprotocol/side/x/dlc/types"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handlePendingOracles(ctx, k)
	handlePendingDCMs(ctx, k)

	generatePriceEventNonces(ctx, k)
	generateDateEventNonces(ctx, k)
	generateLendingEventNonces(ctx, k)
}

// handlePendingOracles handles the pending oracles
func handlePendingOracles(ctx sdk.Context, k keeper.Keeper) {
	pendingOracles := k.GetOracles(ctx, types.DLCOracleStatus_Oracle_Status_Pending)

	for _, oracle := range pendingOracles {
		// check if the pending oracle expired
		if !ctx.BlockTime().Before(oracle.Time.Add(k.DKGTimeoutPeriod(ctx))) {
			oracle.Status = types.DLCOracleStatus_Oracle_Status_Timedout
			k.SetOracle(ctx, oracle)

			continue
		}

		// handle pending pub keys
		pubKeys := k.GetPendingOraclePubKeys(ctx, oracle.Id)
		if len(pubKeys) != len(oracle.Participants) {
			continue
		}

		// check if the pending pub keys are valid
		if !types.CheckPendingPubKeys(pubKeys) {
			oracle.Status = types.DLCOracleStatus_Oracle_Status_Failed
			k.SetOracle(ctx, oracle)

			continue
		}

		// set pub key
		oracle.Pubkey = hex.EncodeToString(pubKeys[0])

		// update status
		oracle.Status = types.DLCOracleStatus_Oracle_status_Enable

		k.SetOracle(ctx, oracle)
		k.SetOracleByPubKey(ctx, oracle.Id, pubKeys[0])
	}
}

// handlePendingDCMs handles the pending DCMs
func handlePendingDCMs(ctx sdk.Context, k keeper.Keeper) {
	pendingDCMs := k.GetDCMs(ctx, types.DCMStatus_DCM_Status_Pending)

	for _, dcm := range pendingDCMs {
		// check if the pending DCM expired
		if !ctx.BlockTime().Before(dcm.Time.Add(k.DKGTimeoutPeriod(ctx))) {
			dcm.Status = types.DCMStatus_DCM_Status_Timedout
			k.SetDCM(ctx, dcm)

			continue
		}

		// handle pending pub keys
		pubKeys := k.GetPendingDCMPubKeys(ctx, dcm.Id)
		if len(pubKeys) != len(dcm.Participants) {
			continue
		}

		// check if the pending pub keys are valid
		if !types.CheckPendingPubKeys(pubKeys) {
			dcm.Status = types.DCMStatus_DCM_Status_Failed
			k.SetDCM(ctx, dcm)

			continue
		}

		// set pub key
		dcm.Pubkey = hex.EncodeToString(pubKeys[0])

		// update status
		dcm.Status = types.DCMStatus_DCM_status_Enable

		k.SetDCM(ctx, dcm)
	}
}

// generatePriceEventNonces emits nonce generation events for dlc price events
func generatePriceEventNonces(ctx sdk.Context, k keeper.Keeper) {
	// get all enabled oracles
	oracles := k.GetOracles(ctx, types.DLCOracleStatus_Oracle_status_Enable)
	if len(oracles) == 0 {
		return
	}

	// select oracle
	selectedOracleId := ctx.BlockHeight() % int64(len(oracles))
	oracle := oracles[selectedOracleId]

	// get price interval and nonce queue size
	priceInterval := int64(k.PriceInterval(ctx, "BTC-USD"))
	nonceQueueSize := int64(k.PriceEventNonceQueueSize(ctx))

	// check if price event nonces need to be generated
	currentPrice := k.GetPrice(ctx, "BTC-USD")
	currentEventPrice := k.GetCurrentEventPrice(ctx, "BTC-USD")
	if currentEventPrice >= currentPrice.Int64()+nonceQueueSize*priceInterval && k.GetTriggeredPriceEventQueueCount(ctx) == 0 {
		return
	}

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGenerateNonce,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d-%d", ctx.BlockHeight(), types.DlcEventType_PRICE)),
			sdk.NewAttribute(types.AttributeKeyDLCEventType, fmt.Sprintf("%d", types.DlcEventType_PRICE)),
			sdk.NewAttribute(types.AttributeKeyOraclePubKey, oracle.Pubkey),
			sdk.NewAttribute(types.AttributeKeyParticipants, strings.Join(oracle.Participants, types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", oracle.Threshold)),
		),
	)
}

// generateDateEventNonces emits nonce generation events for dlc date events
func generateDateEventNonces(ctx sdk.Context, k keeper.Keeper) {
	// get all enabled oracles
	oracles := k.GetOracles(ctx, types.DLCOracleStatus_Oracle_status_Enable)
	if len(oracles) == 0 {
		return
	}

	// select oracle
	selectedOracleId := ctx.BlockHeight() % int64(len(oracles))
	oracle := oracles[selectedOracleId]

	// check if date event nonces need to be generated
	currentEventDate := k.GetCurrentEventDate(ctx)
	if (currentEventDate-ctx.BlockTime().Unix())/k.DateInterval(ctx) >= int64(k.DateEventNonceQueueSize(ctx)) {
		return
	}

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGenerateNonce,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d-%d", ctx.BlockHeight(), types.DlcEventType_DATE)),
			sdk.NewAttribute(types.AttributeKeyDLCEventType, fmt.Sprintf("%d", types.DlcEventType_DATE)),
			sdk.NewAttribute(types.AttributeKeyOraclePubKey, oracle.Pubkey),
			sdk.NewAttribute(types.AttributeKeyParticipants, strings.Join(oracle.Participants, types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", oracle.Threshold)),
		),
	)
}

// generateLendingEventNonces emits nonce generation events for dlc lending events
func generateLendingEventNonces(ctx sdk.Context, k keeper.Keeper) {
	// get all enabled oracles
	oracles := k.GetOracles(ctx, types.DLCOracleStatus_Oracle_status_Enable)
	if len(oracles) == 0 {
		return
	}

	// select oracle
	selectedOracleId := ctx.BlockHeight() % int64(len(oracles))
	oracle := oracles[selectedOracleId]

	// check if lending event nonces need to be generated
	pendingLendingEventCount := k.GetPendingLendingEventCount(ctx)
	if pendingLendingEventCount >= k.LendingEventNonceQueueSize(ctx) {
		return
	}

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGenerateNonce,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d-%d", ctx.BlockHeight(), types.DlcEventType_LENDING)),
			sdk.NewAttribute(types.AttributeKeyDLCEventType, fmt.Sprintf("%d", types.DlcEventType_LENDING)),
			sdk.NewAttribute(types.AttributeKeyOraclePubKey, oracle.Pubkey),
			sdk.NewAttribute(types.AttributeKeyParticipants, strings.Join(oracle.Participants, types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", oracle.Threshold)),
		),
	)
}
