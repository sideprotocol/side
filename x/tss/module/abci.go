package tss

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/tss/keeper"
	"github.com/sideprotocol/side/x/tss/types"
)

// EndBlocker called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleDKGRequests(ctx, k)
	handleRefreshingRequests(ctx, k)
}

// handleDKGRequests performs the DKG request handling
func handleDKGRequests(ctx sdk.Context, k keeper.Keeper) {
	// get pending DKG requests
	pendingDKGRequests := k.GetPendingDKGRequests(ctx)

	for _, req := range pendingDKGRequests {
		// check if the DKG request expired
		if !req.ExpirationTime.IsZero() && !ctx.BlockTime().Before(req.ExpirationTime) {
			req.Status = types.DKGStatus_DKG_STATUS_TIMEDOUT
			k.SetDKGRequest(ctx, req)

			// callback the corresponding module handler
			if err := k.GetDKGRequestTimeoutHandler(req.Module)(ctx, req.Id, req.Type, req.Intent, k.GetAbsentDKGParticipants(ctx, req)); err != nil {
				k.Logger(ctx).Info("Failed to call DGKRequestTimeoutHandler", "module", req.Module, "type", req.Type, "intent", req.Intent)
			}

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

		// callback the corresponding module handler
		if err := k.GetDKGRequestCompletedHandler(req.Module)(ctx, req.Id, req.Type, req.Intent, completions[0].PubKeys); err != nil {
			req.Status = types.DKGStatus_DKG_STATUS_FAILED
			k.SetDKGRequest(ctx, req)

			continue
		}

		// update status
		req.Status = types.DKGStatus_DKG_STATUS_COMPLETED
		k.SetDKGRequest(ctx, req)
	}
}

// handleRefreshingRequests performs the refreshing request handling
func handleRefreshingRequests(ctx sdk.Context, k keeper.Keeper) {
	// get pending refreshing requests
	requests := k.GetPendingRefreshingRequests(ctx)

	for _, req := range requests {
		// check if the refreshing request expired
		if !req.ExpirationTime.IsZero() && !ctx.BlockTime().Before(req.ExpirationTime) {
			req.Status = types.RefreshingStatus_REFRESHING_STATUS_TIMEDOUT
			k.SetRefreshingRequest(ctx, req)

			continue
		}

		// check refreshing completions
		completions := k.GetRefreshingCompletions(ctx, req.Id)
		if len(completions) != len(k.GetRefreshingParticipants(ctx, req)) {
			continue
		}

		// update status
		req.Status = types.RefreshingStatus_REFRESHING_STATUS_COMPLETED
		k.SetRefreshingRequest(ctx, req)

		// update DKG participants and threshold
		dkgRequest := k.GetDKGRequest(ctx, req.DkgId)
		dkgRequest.Participants = k.GetRefreshingParticipants(ctx, req)
		dkgRequest.Threshold = req.Threshold
		k.SetDKGRequest(ctx, dkgRequest)

		// Emit events
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRefreshingCompleted,
				sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", req.Id)),
				sdk.NewAttribute(types.AttributeKeyDKGId, fmt.Sprintf("%d", req.DkgId)),
			),
		)
	}
}
