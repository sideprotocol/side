package lending

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/keeper"
	"github.com/sideprotocol/side/x/lending/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)

	// set pools
	for _, pool := range genState.Pools {
		k.SetPool(ctx, pool)
	}

	// check if the module account exists
	moduleAcc := k.GetModuleAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// check if the repayment escrow account exists
	moduleAcc = k.GetRepaymentEscrowAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.RepaymentEscrowAccount))
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Pools = k.GetAllPools(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
