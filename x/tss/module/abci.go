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
	handleResharingRequests(ctx, k)
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

// handleResharingRequests performs the resharing request handling
func handleResharingRequests(ctx sdk.Context, k keeper.Keeper) {
	// get pending resharing requests
	requests := k.GetPendingResharingRequests(ctx)

	for _, req := range requests {
		// check if the resharing request expired
		if !req.ExpirationTime.IsZero() && !ctx.BlockTime().Before(req.ExpirationTime) {
			req.Status = types.ResharingStatus_RESHARING_STATUS_TIMEDOUT
			k.SetResharingRequest(ctx, req)

			continue
		}

		// check resharing completions
		completions := k.GetResharingCompletions(ctx, req.Id)
		if len(completions) != len(k.GetResharingParticipants(ctx, req)) {
			continue
		}

		// update status
		req.Status = types.ResharingStatus_RESHARING_STATUS_COMPLETED
		k.SetResharingRequest(ctx, req)

		// Emit events
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeResharingCompleted,
				sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", req.Id)),
				sdk.NewAttribute(types.AttributeKeyDKGId, fmt.Sprintf("%d", req.DkgId)),
			),
		)
	}
}
