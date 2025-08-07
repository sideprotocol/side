package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey

	authKeeper      types.AccountKeeper
	bankKeeper      types.BankKeeper
	oracleKeeper    types.OracleKeeper
	tssKeeper       types.TSSKeeper
	btcbridgeKeeper types.BtcBridgeKeeper

	liquidatedDebtHandler types.LiquidatedDebtHandler

	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	oracleKeeper types.OracleKeeper,
	tssKeeper types.TSSKeeper,
	btcbridgeKeeper types.BtcBridgeKeeper,
	authority string,
) *Keeper {
	// ensure the module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	k := &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		memKey:          memKey,
		authKeeper:      authKeeper,
		bankKeeper:      bankKeeper,
		oracleKeeper:    oracleKeeper,
		tssKeeper:       tssKeeper,
		btcbridgeKeeper: btcbridgeKeeper,
		authority:       authority,
	}

	// register signing request completed handler
	tssKeeper.RegisterSigningRequestCompletedHandler(types.ModuleName, k.SigningCompletedHandler)

	return k
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

func (k Keeper) GetModuleAccount(ctx sdk.Context) sdk.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

func (k Keeper) BankKeeper() types.BankKeeper {
	return k.bankKeeper
}

func (k Keeper) OracleKeeper() types.OracleKeeper {
	return k.oracleKeeper
}

func (k Keeper) TSSKeeper() types.TSSKeeper {
	return k.tssKeeper
}

func (k Keeper) BtcBridgeKeeper() types.BtcBridgeKeeper {
	return k.btcbridgeKeeper
}

func (k Keeper) LiquidatedDebtHandler() types.LiquidatedDebtHandler {
	return k.liquidatedDebtHandler
}

func (k *Keeper) SetLiquidatedDebtHandler(handler types.LiquidatedDebtHandler) {
	k.liquidatedDebtHandler = handler
}
