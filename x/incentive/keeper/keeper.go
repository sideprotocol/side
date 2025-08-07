package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/incentive/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey

	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper

	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	authority string,
) Keeper {
	// ensure the module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		authKeeper: authKeeper,
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

func (k Keeper) GetModuleAccount(ctx sdk.Context) sdk.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}
