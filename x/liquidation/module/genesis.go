package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/keeper"
	"github.com/sideprotocol/side/x/liquidation/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)

	// set liquidations
	for _, liquidation := range genState.Liquidations {
		k.SetLiquidation(ctx, liquidation)
	}

	// set liquidation records
	for _, record := range genState.LiquidationRecords {
		k.SetLiquidationRecord(ctx, record)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Params = k.GetParams(ctx)
	genesis.Liquidations = k.GetAllLiquidations(ctx)
	genesis.LiquidationRecords = k.GetAllLiquidationRecords(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
