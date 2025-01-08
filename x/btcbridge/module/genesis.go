package btcbridge

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/keeper"
	"github.com/sideprotocol/side/x/btcbridge/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)

	// set the best block header
	k.SetBestBlockHeader(ctx, genState.BestBlockHeader)
	k.SetBlockHeader(ctx, genState.BestBlockHeader)

	// set block headers
	for _, header := range genState.BlockHeaders {
		k.SetBlockHeader(ctx, header)
	}

	// set utxos
	for _, utxo := range genState.Utxos {
		k.SaveUTXO(ctx, utxo)
	}

	// set dkg request
	if genState.DkgRequest != nil {
		k.SetDKGRequest(ctx, genState.DkgRequest)
		k.SetDKGRequestID(ctx, genState.DkgRequest.Id)
	}

	// sort vaults and set the latest vault version
	if len(genState.Params.Vaults) > 0 {
		vaults := genState.Params.Vaults
		sort.Slice(vaults, func(i, j int) bool { return vaults[i].Version < vaults[j].Version })

		k.SetVaultVersion(ctx, vaults[len(vaults)-1].Version)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.BestBlockHeader = k.GetBestBlockHeader(ctx)
	genesis.BlockHeaders = k.GetAllBlockHeaders(ctx)
	genesis.Utxos = k.GetAllUTXOs(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
