package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/tss/keeper"
	"github.com/sideprotocol/side/x/tss/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)

	// set DKG requests
	for _, req := range genState.DkgRequests {
		k.SetDKGRequest(ctx, req)
	}

	// set signing requests
	for _, req := range genState.SigningRequests {
		k.SetSigningRequest(ctx, req)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Params = k.GetParams(ctx)
	genesis.DkgRequests = k.GetDKGRequests(ctx, types.DKGStatus_DKG_STATUS_PENDING)
	genesis.SigningRequests = k.GetSigningRequests(ctx, types.SigningStatus_SIGNING_STATUS_PENDING)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
