package incentive_test

// Path: x/incentive/genesis_test.go

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/sideprotocol/side/testutil/keeper"
	"github.com/sideprotocol/side/testutil/nullify"
	incentive "github.com/sideprotocol/side/x/incentive/module"
	"github.com/sideprotocol/side/x/incentive/types"
)

func TestGenesis(t *testing.T) {
	mnemonic := "sunny bamboo garlic fold reopen exile letter addict forest vessel square lunar shell number deliver cruise calm artist fire just kangaroo suit wheel extend"
	println(mnemonic)

	genesisState := types.DefaultGenesis()

	k, ctx := keepertest.IncentiveKeeper(t)
	incentive.InitGenesis(ctx, k, *genesisState)
	got := incentive.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
