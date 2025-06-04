package v2

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	lendingtypes "github.com/sideprotocol/side/x/lending/types"
	liquidationtypes "github.com/sideprotocol/side/x/liquidation/types"
	oracletypes "github.com/sideprotocol/side/x/oracle/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// UpgradeName is the upgrade version name
const UpgradeName = "v2.0.0"

var StoreUpgrades = storetypes.StoreUpgrades{
	Added: []string{
		tsstypes.ModuleName,
		oracletypes.ModuleName,
		dlctypes.ModuleName,
		lendingtypes.ModuleName,
		liquidationtypes.ModuleName,
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
