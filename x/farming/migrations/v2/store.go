package v2

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/farming/types"
)

// MigrateStore migrates the x/farming module state from the consensus version 1 to
// version 2
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateStakings(ctx, storeKey, cdc)
	migrateCurrentEpochStakingQueue(ctx, storeKey, cdc)

	return nil
}

// migrateStakings performs the stakings migration
func migrateStakings(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.StakingKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// unmarshal staking to v1
		var stakingV1 types.StakingV1
		cdc.MustUnmarshal(iterator.Value(), &stakingV1)

		// build new staking
		staking := &types.Staking{
			Id:              stakingV1.Id,
			Address:         stakingV1.Address,
			Amount:          stakingV1.Amount,
			LockDuration:    stakingV1.LockDuration,
			LockMultiplier:  stakingV1.LockMultiplier,
			EffectiveAmount: stakingV1.EffectiveAmount,
			PendingRewards:  stakingV1.PendingRewards,
			TotalRewards:    stakingV1.PendingRewards,
			StartTime:       stakingV1.StartTime,
			Status:          stakingV1.Status,
		}

		// update staking
		store.Set(iterator.Key(), cdc.MustMarshal(staking))
	}
}

// migrateCurrentEpochStakingQueue performs the staking queue migration for the current epoch
func migrateCurrentEpochStakingQueue(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.CurrentEpochStakingQueueKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		stakingId := sdk.BigEndianToUint64(iterator.Key()[1:])

		bz := store.Get(types.StakingKey(stakingId))
		var staking types.Staking
		cdc.MustUnmarshal(bz, &staking)

		// add to staking queue by address
		store.Set(types.CurrentEpochStakingQueueByAddressKey(staking.Address, stakingId), []byte{})
	}
}
