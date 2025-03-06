package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/oracle/types"
)

func (k Keeper) HasBlockHeader(ctx sdk.Context, hash string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BitcoinHeaderKey(hash))
}

func (k Keeper) SetBlockHeader(ctx sdk.Context, header *types.BlockHeader) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(header)
	store.Set(types.BitcoinHeaderKey(header.Hash), bz)
}

func (k Keeper) GetBlockHeader(ctx sdk.Context, hash string) types.BlockHeader {
	store := ctx.KVStore(k.storeKey)
	var header types.BlockHeader
	bz := store.Get(types.BitcoinHeaderKey(hash))
	k.cdc.MustUnmarshal(bz, &header)
	return header
}
