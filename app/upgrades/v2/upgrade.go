package v2

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	incentivetypes "github.com/sideprotocol/side/x/incentive/types"
)

// UpgradeName is the upgrade version name
const UpgradeName = "v2"

var StoreUpgrades = storetypes.StoreUpgrades{
	Added: []string{
		incentivetypes.ModuleName,
	},
}

// CreateUpgradeHandler creates the upgrade handler
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
