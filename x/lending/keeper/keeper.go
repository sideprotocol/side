package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"

	"github.com/sideprotocol/side/x/lending/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		authKeeper        types.AccountKeeper
		bankKeeper        types.BankKeeper
		mintKeeper        mintkeeper.Keeper
		oracleKeeper      types.OracleKeeper
		liquidationKeeper types.LiquidationKeeper
		dlcKeeper         types.DLCKeeper
		btcbridgeKeeper   types.BtcBridgeKeeper
		tssKeeper         types.TSSKeeper

		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ak types.AccountKeeper,
	bankKeeper types.BankKeeper,
	mintKeeper mintkeeper.Keeper,
	oracleKeeper types.OracleKeeper,
	liquidationKeeper types.LiquidationKeeper,
	dlcKeeper types.DLCKeeper,
	btcbridgeKeeper types.BtcBridgeKeeper,
	tssKeeper types.TSSKeeper,
	authority string,
) Keeper {
	// ensure escrow module account is set
	if addr := ak.GetModuleAddress(types.RepaymentEscrowAccount); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.RepaymentEscrowAccount))
	}

	k := Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		memKey:            memKey,
		authKeeper:        ak,
		bankKeeper:        bankKeeper,
		mintKeeper:        mintKeeper,
		oracleKeeper:      oracleKeeper,
		liquidationKeeper: liquidationKeeper,
		dlcKeeper:         dlcKeeper,
		btcbridgeKeeper:   btcbridgeKeeper,
		tssKeeper:         tssKeeper,
		authority:         authority,
	}

	// set liquidated debt handler for liquidation
	liquidationKeeper.SetLiquidatedDebtHandler(k.HandleLiquidatedDebt)

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

func (k Keeper) GetBlocksPerYear(ctx sdk.Context) uint64 {
	params, err := k.mintKeeper.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	return params.BlocksPerYear
}

func (k Keeper) OracleKeeper() types.OracleKeeper {
	return k.oracleKeeper
}

func (k Keeper) LiquidationKeeper() types.LiquidationKeeper {
	return k.liquidationKeeper
}

func (k Keeper) DLCKeeper() types.DLCKeeper {
	return k.dlcKeeper
}

func (k Keeper) BtcBridgeKeeper() types.BtcBridgeKeeper {
	return k.btcbridgeKeeper
}

func (k Keeper) TSSKeeper() types.TSSKeeper {
	return k.tssKeeper
}
