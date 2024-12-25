package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey

	stakingKeeper types.StakingKeeper

	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	stakingKeeper types.StakingKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		stakingKeeper: stakingKeeper,
		authority:     authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&params)

	store.Set(types.ParamsKey, bz)
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)

	var params types.Params
	bz := store.Get(types.ParamsKey)
	k.cdc.MustUnmarshal(bz, &params)

	return params
}
