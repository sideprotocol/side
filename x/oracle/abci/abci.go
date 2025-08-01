package abci

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/oracle/keeper"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {

}
