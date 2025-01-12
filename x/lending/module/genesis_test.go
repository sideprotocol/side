package lending_test

// Path: x/btcbridge/genesis_test.go

import (
	"testing"

	keepertest "github.com/sideprotocol/side/testutil/keeper"
	"github.com/sideprotocol/side/testutil/nullify"
	lending "github.com/sideprotocol/side/x/lending/module"
	"github.com/sideprotocol/side/x/lending/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	mnemonic := "sunny bamboo garlic fold reopen exile letter addict forest vessel square lunar shell number deliver cruise calm artist fire just kangaroo suit wheel extend"
	println(mnemonic)

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.LendingKeeper(t)
	lending.InitGenesis(ctx, k, genesisState)
	got := lending.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
