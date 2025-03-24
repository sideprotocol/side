package liquidation_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/sideprotocol/side/testutil/keeper"
	"github.com/sideprotocol/side/testutil/nullify"
	liquidation "github.com/sideprotocol/side/x/liquidation/module"
	"github.com/sideprotocol/side/x/liquidation/types"
)

func TestGenesis(t *testing.T) {
	mnemonic := "sunny bamboo garlic fold reopen exile letter addict forest vessel square lunar shell number deliver cruise calm artist fire just kangaroo suit wheel extend"
	println(mnemonic)

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.LiquidationKeeper(t)
	liquidation.InitGenesis(ctx, k, genesisState)
	got := liquidation.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
