package dlc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/keeper"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handlePendingOracles(ctx, k)
	handlePendingAgencies(ctx, k)
}

// handlePendingOracles handles the pending oracles
func handlePendingOracles(ctx sdk.Context, k keeper.Keeper) {

}

// handlePendingAgencies handles the pending agencies
func handlePendingAgencies(ctx sdk.Context, k keeper.Keeper) {

}
