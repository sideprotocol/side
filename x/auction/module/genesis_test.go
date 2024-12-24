package auction_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/sideprotocol/side/testutil/keeper"
	"github.com/sideprotocol/side/testutil/nullify"
	auction "github.com/sideprotocol/side/x/auction/module"
	"github.com/sideprotocol/side/x/auction/types"
)

func TestGenesis(t *testing.T) {
	mnemonic := "sunny bamboo garlic fold reopen exile letter addict forest vessel square lunar shell number deliver cruise calm artist fire just kangaroo suit wheel extend"
	println(mnemonic)

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.AuctionKeeper(t)
	auction.InitGenesis(ctx, k, genesisState)
	got := auction.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
