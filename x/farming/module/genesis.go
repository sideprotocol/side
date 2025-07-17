package farming

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/farming/keeper"
	"github.com/sideprotocol/side/x/farming/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)

	// set stakings
	for _, staking := range genState.Stakings {
		k.SetStaking(ctx, staking)
	}

	// start the new epoch if farming enabled
	if genState.Params.Enabled {
		k.NewEpoch(ctx)
	}

	// check if the module account exists
	moduleAcc := k.GetModuleAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set the module account if there is zero balance
	balances := k.BankKeeper().GetAllBalances(ctx, moduleAcc.GetAddress())
	if balances.IsZero() {
		k.AuthKeeper().SetModuleAccount(ctx, moduleAcc)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Params = k.GetParams(ctx)
	genesis.Stakings = k.GetAllStakings(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
