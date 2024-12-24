package dlc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/keeper"
	"github.com/sideprotocol/side/x/dlc/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)

	// set events
	for _, event := range genState.Events {
		k.SetEvent(ctx, event)
	}

	// set attestations
	for _, attestation := range genState.Attestations {
		k.SetAttestation(ctx, attestation)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Params = k.GetParams(ctx)
	genesis.Events = k.GetAllEvents(ctx)
	genesis.Attestations = k.GetAttestations(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
