package keeper

import (
	"fmt"

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

		authKeeper      types.AccountKeeper
		bankKeeper      types.BankKeeper
		oracleKeeper    types.OracleKeeper
		auctionKeeper   types.AuctionKeeper
		dlcKeeper       types.DLCKeeper
		btcbridgeKeeper types.BtcBridgeKeeper

		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ak types.AccountKeeper,
	bankKeeper types.BankKeeper,
	oracleKeeper types.OracleKeeper,
	auctionKeeper types.AuctionKeeper,
	dlcKeeper types.DLCKeeper,
	btcbridgeKeeper types.BtcBridgeKeeper,
	authority string,
) Keeper {
	// ensure escrow module account is set
	if addr := ak.GetModuleAddress(types.RepaymentEscrowAccount); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.RepaymentEscrowAccount))
	}

	return Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		memKey:          memKey,
		authKeeper:      ak,
		bankKeeper:      bankKeeper,
		oracleKeeper:    oracleKeeper,
		auctionKeeper:   auctionKeeper,
		dlcKeeper:       dlcKeeper,
		btcbridgeKeeper: btcbridgeKeeper,
		authority:       authority,
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

func (k Keeper) OracleKeeper() types.OracleKeeper {
	return k.oracleKeeper
}

func (k Keeper) AuctionKeeper() types.AuctionKeeper {
	return k.auctionKeeper
}

func (k Keeper) DLCKeeper() types.DLCKeeper {
	return k.dlcKeeper
}

func (k Keeper) BtcBridgeKeeper() types.BtcBridgeKeeper {
	return k.btcbridgeKeeper
}
