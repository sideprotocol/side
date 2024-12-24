package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/auction/keeper"
	"github.com/sideprotocol/side/x/auction/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)

	// set auctions
	for _, auction := range genState.Auctions {
		k.SetAuction(ctx, auction)
	}

	// set bids
	for _, bid := range genState.Bids {
		k.SetBid(ctx, bid)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.Auctions = k.GetAllAuctions(ctx)
	genesis.Bids = k.GetAllBids(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
