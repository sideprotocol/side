package liquidation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/keeper"
)

// EndBlocker called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {

}
