package v3

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// MigrateStore migrates the x/dlc module state from the consensus version 1 to
// version 2
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateParams(ctx, storeKey, cdc)
	migrateStore(ctx, storeKey, cdc)

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

	for _, participant := range params.AllowedOracleParticipants {
		liveness := &types.OracleParticipantLiveness{
			ConsensusPubkey: participant,
			IsAlive:         true,
			LastDkgId:       0,
			LastBlockHeight: 0,
		}

		bz := cdc.MustMarshal(liveness)

		store.Set(types.OracleParticipantLivenessKey(participant), bz)
	}
}

// migrateStore performs the store migration
func migrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	// get the current dcm id
	dcmIdBz := store.Get(types.DCMIdKey)

	// update the current oracle id to the latest dcm id
	store.Set(types.OracleIdKey, dcmIdBz)

	// update the current dcm id to the latest dcm id
	iterateDCMs(ctx, storeKey, cdc, func(dcm *types.DCM) (stop bool) {
		store.Set(types.DCMIdKey, sdk.Uint64ToBigEndian(dcm.Id))
		return false
	})
}

// iterateDCMs iterates through all DCMs
func iterateDCMs(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, cb func(dcm *types.DCM) (stop bool)) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.DCMKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var dcm types.DCM
		cdc.MustUnmarshal(iterator.Value(), &dcm)

		if cb(&dcm) {
			break
		}
	}
}
