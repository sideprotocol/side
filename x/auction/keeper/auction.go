package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/auction/types"
)

// GetAuctionId gets the current auction id
func (k Keeper) GetAuctionId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.AuctionIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementAuctionId increments the auction id and returns the new id
func (k Keeper) IncrementAuctionId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetAuctionId(ctx) + 1
	store.Set(types.AuctionIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// HasAuction returns true if the given auction exists, false otherwise
func (k Keeper) HasAuction(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.AuctionKey(id))
}

// GetAuction gets the auction by the given id
func (k Keeper) GetAuction(ctx sdk.Context, id uint64) *types.Auction {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.AuctionKey(id))
	var auction types.Auction
	k.cdc.MustUnmarshal(bz, &auction)

	return &auction
}

// SetAuction sets the given auction
func (k Keeper) SetAuction(ctx sdk.Context, auction *types.Auction) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(auction)
	store.Set(types.AuctionKey(auction.Id), bz)
}

// GetAllAuctions gets all auctions
func (k Keeper) GetAllAuctions(ctx sdk.Context) []*types.Auction {
	auctions := make([]*types.Auction, 0)

	k.IterateAuctions(ctx, func(auction *types.Auction) (stop bool) {
		auctions = append(auctions, auction)
		return false
	})

	return auctions
}

// GetAuctions gets auctions by the given status
func (k Keeper) GetAuctions(ctx sdk.Context, status types.AuctionStatus) []*types.Auction {
	auctions := make([]*types.Auction, 0)

	k.IterateAuctions(ctx, func(auction *types.Auction) (stop bool) {
		if auction.Status == status {
			auctions = append(auctions, auction)
		}

		return false
	})

	return auctions
}

// IterateAuctions iterates through all auctions
func (k Keeper) IterateAuctions(ctx sdk.Context, cb func(auction *types.Auction) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.AuctionKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var auction types.Auction
		k.cdc.MustUnmarshal(iterator.Value(), &auction)

		if cb(&auction) {
			break
		}
	}
}
