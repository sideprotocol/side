package v093

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// UpgradeName is the upgrade version name
const UpgradeName = "v0.9.3"

// CreateUpgradeHandler creates the upgrade handler
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
