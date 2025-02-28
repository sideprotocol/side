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
	handlePendingAgencies(ctx, k)
	generateNonces(ctx, k)
}

// handlePendingOracles handles the pending oracles
func handlePendingOracles(ctx sdk.Context, k keeper.Keeper) {
	pendingOracles := k.GetOracles(ctx, types.DLCOracleStatus_Oracle_Status_Pending)

	for _, oracle := range pendingOracles {
		// check if the pending oracle expired
		if !ctx.BlockTime().Before(oracle.Time.Add(k.GetDKGTimeoutPeriod(ctx))) {
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

// handlePendingAgencies handles the pending agencies
func handlePendingAgencies(ctx sdk.Context, k keeper.Keeper) {
	pendingAgencies := k.GetAgencies(ctx, types.AgencyStatus_Agency_Status_Pending)

	for _, agency := range pendingAgencies {
		// check if the pending agency expired
		if !ctx.BlockTime().Before(agency.Time.Add(k.GetDKGTimeoutPeriod(ctx))) {
			agency.Status = types.AgencyStatus_Agency_Status_Timedout
			k.SetAgency(ctx, agency)

			continue
		}

		// handle pending pub keys
		pubKeys := k.GetPendingAgencyPubKeys(ctx, agency.Id)
		if len(pubKeys) != len(agency.Participants) {
			continue
		}

		// check if the pending pub keys are valid
		if !types.CheckPendingPubKeys(pubKeys) {
			agency.Status = types.AgencyStatus_Agency_Status_Failed
			k.SetAgency(ctx, agency)

			continue
		}

		// set pub key
		agency.Pubkey = hex.EncodeToString(pubKeys[0])

		// update status
		agency.Status = types.AgencyStatus_Agency_status_Enable

		k.SetAgency(ctx, agency)
	}
}

// generateNonces emits nonce generation events
func generateNonces(ctx sdk.Context, k keeper.Keeper) {
	// get all enabled oracles
	oracles := k.GetOracles(ctx, types.DLCOracleStatus_Oracle_status_Enable)
	if len(oracles) == 0 {
		return
	}

	// select oralce
	selectedOracleId := ctx.BlockHeight() % int64(len(oracles))
	oracle := oracles[selectedOracleId]

	// get nonce index and params
	nonceIndex := k.GetNonceIndex(ctx, oracle.Id)
	nonceQueueSize := uint64(k.GetNonceQueueSize(ctx))

	// check if nonces need to be generated
	currentPrice := k.GetPrice(ctx, "BTC-USD")
	currentEventPrice := k.GetCurrentEventPrice(ctx, "BTC-USD")
	if currentEventPrice > 0 && currentEventPrice >= currentPrice.Int64() && nonceIndex >= nonceQueueSize {
		return
	}

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGenerateNonce,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", nonceIndex+1)),
			sdk.NewAttribute(types.AttributeKeyOraclePubKey, oracle.Pubkey),
			sdk.NewAttribute(types.AttributeKeyParticipants, strings.Join(oracle.Participants, types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", oracle.Threshold)),
		),
	)
}
