package v3

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// MigrateStore migrates the x/dlc module state from the consensus version 2 to
// version 3
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateParams(ctx, storeKey, cdc)

	return nil
}

// migrateParams performs the params migration
func migrateParams(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	// get current params
	var paramsV1 types.ParamsV1
	bz := store.Get(types.ParamsKey)
	cdc.MustUnmarshal(bz, &paramsV1)

	// build new params
	params := &types.Params{
		NonceQueueSize:             paramsV1.LendingEventNonceQueueSize,
		NonceGenerationBatchSize:   paramsV1.NonceGenerationBatchSize,
		NonceGenerationInterval:    paramsV1.NonceGenerationInterval,
		AllowedOracleParticipants:  paramsV1.AllowedOracleParticipants,
		OracleParticipantNum:       paramsV1.OracleParticipantNum,
		OracleParticipantThreshold: paramsV1.OracleParticipantThreshold,
	}

	bz = cdc.MustMarshal(params)
	store.Set(types.ParamsKey, bz)
}
