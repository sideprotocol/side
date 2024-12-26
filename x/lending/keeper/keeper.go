package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/lending/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		bankKeeper types.BankKeeper

		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		bankKeeper: bankKeeper,
		authority:  authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsStoreKey, bz)
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	var params types.Params
	bz := store.Get(types.ParamsStoreKey)
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

func (k Keeper) SetPool(ctx sdk.Context, pool types.LendingPool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&pool)
	store.Set(types.PoolStoreKey(pool.Id), bz)
}

func (k Keeper) HasPool(ctx sdk.Context, pool_id string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.PoolStoreKey(pool_id))
}

func (k Keeper) GetPool(ctx sdk.Context, pool_id string) types.LendingPool {
	store := ctx.KVStore(k.storeKey)
	var pool types.LendingPool
	bz := store.Get(types.PoolStoreKey(pool_id))
	k.cdc.MustUnmarshal(bz, &pool)
	return pool
}

// GetAllBlockHeaders returns all block headers
func (k Keeper) GetAllPools(ctx sdk.Context) []*types.LendingPool {
	var pools []*types.LendingPool
	k.IteratePools(ctx, func(pool types.LendingPool) (stop bool) {
		pools = append(pools, &pool)
		return false
	})
	return pools
}

// IterateBlockHeaders iterates through all block headers
func (k Keeper) IteratePools(ctx sdk.Context, process func(header types.LendingPool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PoolStorePrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var header types.LendingPool
		k.cdc.MustUnmarshal(iterator.Value(), &header)
		if process(header) {
			break
		}
	}
}
