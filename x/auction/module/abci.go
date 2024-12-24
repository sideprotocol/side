package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/auction/keeper"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleAuction(ctx, k)
}

// handleAuction performs the auction handling
func handleAuction(ctx sdk.Context, k keeper.Keeper) {

}
