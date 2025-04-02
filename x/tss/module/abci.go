package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/tss/keeper"
	"github.com/sideprotocol/side/x/tss/types"
)

// EndBlocker called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleDKGRequests(ctx, k)
}

// handleDKGRequests performs the DKG request handling
func handleDKGRequests(ctx sdk.Context, k keeper.Keeper) {
	// get pending DKG requests
	pendingDKGRequests := k.GetPendingDKGRequests(ctx)

	for _, req := range pendingDKGRequests {
		// check if the DKG request expired
		if !ctx.BlockTime().Before(req.ExpirationTime) {
			req.Status = types.DKGStatus_DKG_STATUS_TIMEDOUT
			k.SetDKGRequest(ctx, req)

			continue
		}

		// get DKG completions
		completions := k.GetDKGCompletions(ctx, req.Id)
		if len(completions) != len(req.Participants) {
			continue
		}

		// check if the DKG completions are valid
		if !types.CheckDKGCompletions(completions) {
			req.Status = types.DKGStatus_DKG_STATUS_FAILED
			k.SetDKGRequest(ctx, req)

			continue
		}

		// update status
		req.Status = types.DKGStatus_DKG_STATUS_COMPLETED
		k.SetDKGRequest(ctx, req)
	}
}
