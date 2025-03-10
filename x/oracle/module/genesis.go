package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/oracle/keeper"
	"github.com/sideprotocol/side/x/oracle/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)

	// set block headers
	k.SetBlockHeaders(ctx, genState.Blocks)

	// set oracle prices
	for _, op := range genState.Prices {
		k.SetPrice(ctx, op.Symbol, op.Price.String())
	}

}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Params = k.GetParams(ctx)
	// genesis.Blocks = k.GetBlockHeaders()
	// genesis.Prices = k.GetPrices(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
