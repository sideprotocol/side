package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/sideprotocol/side/x/dlc/migrations/v2"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// Migrator is a struct for handling in-place store migrations
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, m.keeper.storeKey, storetypes.NewKVStoreKey(tsstypes.StoreKey), m.keeper.cdc)
}
