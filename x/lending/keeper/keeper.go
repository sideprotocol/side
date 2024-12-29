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

// GetAllPools returns all block headers
func (k Keeper) GetAllPools(ctx sdk.Context) []*types.LendingPool {
	var pools []*types.LendingPool
	k.IteratePools(ctx, func(pool types.LendingPool) (stop bool) {
		pools = append(pools, &pool)
		return false
	})
	return pools
}

// IteratePools iterates through all block headers
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

func (k Keeper) SetLoan(ctx sdk.Context, loan types.Loan) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&loan)
	store.Set(types.LoanStoreKey(loan.VaultAddress), bz)
}

func (k Keeper) HasLoan(ctx sdk.Context, vault string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.LoanStoreKey(vault))
}

func (k Keeper) GetLoan(ctx sdk.Context, vault string) types.Loan {
	store := ctx.KVStore(k.storeKey)
	var loan types.Loan
	bz := store.Get(types.LoanStoreKey(vault))
	k.cdc.MustUnmarshal(bz, &loan)
	return loan
}

// IterateLoans iterates through all block headers
func (k Keeper) IterateLoans(ctx sdk.Context, process func(header types.Loan) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.LoanStorePrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var header types.Loan
		k.cdc.MustUnmarshal(iterator.Value(), &header)
		if process(header) {
			break
		}
	}
}

// GetAllLoans returns all block headers
func (k Keeper) GetAllLoans(ctx sdk.Context) []*types.Loan {
	var loans []*types.Loan
	k.IterateLoans(ctx, func(loan types.Loan) (stop bool) {
		loans = append(loans, &loan)
		return false
	})
	return loans
}
