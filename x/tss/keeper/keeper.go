package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/tss/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey

	stakingKeeper types.StakingKeeper

	dkgCompletionReceivedHandlers   map[string]types.DKGCompletionReceivedHandler
	dkgRequestCompletedHandlers     map[string]types.DKGRequestCompletedHandler
	dkgRequestTimeoutHandlers       map[string]types.DKGRequestTimeoutHandler
	signingRequestCompletedHandlers map[string]types.SigningRequestCompletedHandler

	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	stakingKeeper types.StakingKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:                             cdc,
		storeKey:                        storeKey,
		memKey:                          memKey,
		stakingKeeper:                   stakingKeeper,
		dkgCompletionReceivedHandlers:   make(map[string]types.DKGCompletionReceivedHandler),
		dkgRequestCompletedHandlers:     make(map[string]types.DKGRequestCompletedHandler),
		dkgRequestTimeoutHandlers:       make(map[string]types.DKGRequestTimeoutHandler),
		signingRequestCompletedHandlers: make(map[string]types.SigningRequestCompletedHandler),
		authority:                       authority,
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

func (k *Keeper) RegisterDKGCompletionReceivedHandler(module string, handler types.DKGCompletionReceivedHandler) {
	k.dkgCompletionReceivedHandlers[module] = handler
}

func (k *Keeper) RegisterDKGRequestCompletedHandler(module string, handler types.DKGRequestCompletedHandler) {
	k.dkgRequestCompletedHandlers[module] = handler
}

func (k *Keeper) RegisterDKGRequestTimeoutHandler(module string, handler types.DKGRequestTimeoutHandler) {
	k.dkgRequestTimeoutHandlers[module] = handler
}

func (k *Keeper) RegisterSigningRequestCompletedHandler(module string, handler types.SigningRequestCompletedHandler) {
	k.signingRequestCompletedHandlers[module] = handler
}

func (k Keeper) GetDKGCompletionReceivedHandler(module string) types.DKGCompletionReceivedHandler {
	return k.dkgCompletionReceivedHandlers[module]
}

func (k Keeper) GetDKGRequestCompletedHandler(module string) types.DKGRequestCompletedHandler {
	return k.dkgRequestCompletedHandlers[module]
}

func (k Keeper) GetDKGRequestTimeoutHandler(module string) types.DKGRequestTimeoutHandler {
	return k.dkgRequestTimeoutHandlers[module]
}

func (k Keeper) GetSigningRequestCompletedHandler(module string) types.SigningRequestCompletedHandler {
	return k.signingRequestCompletedHandlers[module]
}
