package keeper

import (
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/oracle/types"
)

func (k Keeper) HasBlockHeader(ctx sdk.Context, hash string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BitcoinHeaderKey(hash))
}

func (k Keeper) SetBlockHeaders(ctx sdk.Context, headers []*types.BlockHeader) error {
	// store := ctx.KVStore(k.storeKey)
	length := len(headers)
	if length == 0 {
		return nil
	}

	slices.SortFunc(headers, func(a *types.BlockHeader, b *types.BlockHeader) int { return (int)(a.Height - b.Height) })

	best := k.GetBestBlockHeader(ctx)
	for _, h := range headers {
		if best.Hash != h.PreviousBlockHash {
			return types.ErrInvalidBlockHeaders
		}
		k.SetBlockHeader(ctx, h)
	}

	k.SetBestBlockHeader(ctx, headers[length-1])

	return nil

}

func (k Keeper) SetBlockHeader(ctx sdk.Context, header *types.BlockHeader) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(header)
	store.Set(types.BitcoinHeaderKey(header.Hash), bz)
	store.Set(types.BitcoinBlockHeaderHeightKey(header.Height), []byte(header.Hash))
}

func (k Keeper) GetBlockHeader(ctx sdk.Context, hash string) *types.BlockHeader {
	store := ctx.KVStore(k.storeKey)
	var header types.BlockHeader
	bz := store.Get(types.BitcoinHeaderKey(hash))
	k.cdc.MustUnmarshal(bz, &header)
	return &header
}

func (k Keeper) GetBestBlockHeader(ctx sdk.Context) *types.BlockHeader {
	store := ctx.KVStore(k.storeKey)
	var blockHeader types.BlockHeader
	bz := store.Get(types.BitcoinBestBlockHeaderKey)
	k.cdc.MustUnmarshal(bz, &blockHeader)
	return &blockHeader
}

func (k Keeper) SetBestBlockHeader(ctx sdk.Context, header *types.BlockHeader) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(header)
	store.Set(types.BitcoinBestBlockHeaderKey, bz)
}

func (k Keeper) GetBlockHashByHeight(ctx sdk.Context, height int32) string {
	store := ctx.KVStore(k.storeKey)
	hash := store.Get(types.BitcoinBlockHeaderHeightKey(height))
	return string(hash)
}

func (k Keeper) GetBlockHeaderByHeight(ctx sdk.Context, height int32) *types.BlockHeader {
	store := ctx.KVStore(k.storeKey)
	hash := store.Get(types.BitcoinBlockHeaderHeightKey(height))
	return k.GetBlockHeader(ctx, string(hash))
}
