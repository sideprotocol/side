package v2_rc10

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	farmingtypes "github.com/sideprotocol/side/x/farming/types"
)

// UpgradeName is the upgrade version name
const UpgradeName = "v2.0.0-rc.10"

var StoreUpgrades = storetypes.StoreUpgrades{
	Added: []string{
		farmingtypes.ModuleName,
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
