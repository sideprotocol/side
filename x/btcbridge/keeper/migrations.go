package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/sideprotocol/side/x/btcbridge/migrations/v2"
	v3 "github.com/sideprotocol/side/x/btcbridge/migrations/v3"
	v4 "github.com/sideprotocol/side/x/btcbridge/migrations/v4"
	v5 "github.com/sideprotocol/side/x/btcbridge/migrations/v5"
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
	return v2.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc)
}

// Migrate2to3 migrates from version 2 to 3
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc)
}

// Migrate3to4 migrates from version 3 to 4
func (m Migrator) Migrate3to4(ctx sdk.Context) error {
	return v4.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc)
}

// Migrate4to5 migrates from version 4 to 5
func (m Migrator) Migrate4to5(ctx sdk.Context) error {
	return v5.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc, m.keeper.bankKeeper)
}
