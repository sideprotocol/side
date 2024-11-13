package v095

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// UpgradeName is the upgrade version name
const UpgradeName = "v0.9.5"

// CreateUpgradeHandler creates the upgrade handler
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	paramsKeeper paramskeeper.Keeper,
	consensusParamsKeeper consensusparamkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	baseAppLegacySS := paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		baseapp.MigrateParams(sdk.UnwrapSDKContext(ctx), baseAppLegacySS, consensusParamsKeeper.ParamsStore)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
