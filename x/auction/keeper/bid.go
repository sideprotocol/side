package keeper

import (
	"sort"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/auction/types"
)

// HandleBid performs the bid handling
func (k Keeper) HandleBid(ctx sdk.Context, sender string, auctionId uint64, price int64, amount sdk.Coin) (*types.Bid, error) {
	if !k.HasAuction(ctx, auctionId) {
		return nil, types.ErrAuctionDoesNotExist
	}

	auction := k.GetAuction(ctx, auctionId)
	if auction.Status != types.AuctionStatus_AUCTION_STATUS_OPEN {
		return nil, types.ErrAuctionEnded
	}

	if amount.Amount.Uint64() < k.GetParams(ctx).MinBidAmount {
		return nil, errorsmod.Wrap(types.ErrInvalidBid, "amount can not be less than the minimum allowed amount")
	}

	bidValue := sdk.NewInt64Coin("uusdc", price*amount.Amount.Int64())
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(sender), types.ModuleName, sdk.NewCoins(bidValue)); err != nil {
		return nil, err
	}

	bid := &types.Bid{
		Id:        k.IncrementBidId(ctx),
		Bidder:    sender,
		AuctionId: auctionId,
		BidPrice:  price,
		BidAmount: amount,
		Status:    types.BidStatus_BID_STATUS_BIDDING,
	}

	k.SetBid(ctx, bid)

	return bid, nil
}

// CancelBid cancels the specified bid
func (k Keeper) CancelBid(ctx sdk.Context, sender string, id uint64) error {
	if !k.HasBid(ctx, id) {
		return types.ErrBidDoesNotExist
	}

	bid := k.GetBid(ctx, id)
	if bid.Bidder != sender {
		return errorsmod.Wrap(types.ErrUnauthorized, "sender is not the bidder")
	}

	if bid.Status != types.BidStatus_BID_STATUS_BIDDING {
		return types.ErrInvalidBidStatus
	}

	bidValue := sdk.NewInt64Coin("uusdc", bid.BidPrice*bid.BidAmount.Amount.Int64())
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(bidValue)); err != nil {
		return err
	}

	bid.Status = types.BidStatus_BID_STATUS_CANCELLED

	k.SetBid(ctx, bid)

	return nil
}

// GetBidId gets the current bid id
func (k Keeper) GetBidId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.BidIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementBidId increments the bid id and returns the new id
func (k Keeper) IncrementBidId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetBidId(ctx) + 1
	store.Set(types.BidIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// HasBid returns true if the given bid exists, false otherwise
func (k Keeper) HasBid(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.BidKey(id))
}

// GetBid gets the bid by the given id
func (k Keeper) GetBid(ctx sdk.Context, id uint64) *types.Bid {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.BidKey(id))
	var bid types.Bid
	k.cdc.MustUnmarshal(bz, &bid)

	return &bid
}

// SetBid sets the given bid
func (k Keeper) SetBid(ctx sdk.Context, bid *types.Bid) {
	store := ctx.KVStore(k.storeKey)

	k.SetBidByAuction(ctx, bid)

	bz := k.cdc.MustMarshal(bid)
	store.Set(types.BidKey(bid.Id), bz)
}

// SetBidByAuction sets the given bid by auction
func (k Keeper) SetBidByAuction(ctx sdk.Context, bid *types.Bid) {
	store := ctx.KVStore(k.storeKey)

	if k.HasBid(ctx, bid.Id) {
		previousStatus := k.GetBid(ctx, bid.Id).Status
		store.Delete(types.BidByAuctionKey(bid.AuctionId, bid.Id, previousStatus))
	}

	store.Set(types.BidByAuctionKey(bid.AuctionId, bid.Id, bid.Status), []byte{})
}

// GetAllBids gets all bids
func (k Keeper) GetAllBids(ctx sdk.Context) []*types.Bid {
	bids := make([]*types.Bid, 0)

	k.IterateBids(ctx, func(bid *types.Bid) (stop bool) {
		bids = append(bids, bid)
		return false
	})

	return bids
}

// GetBids gets bids by the given status
func (k Keeper) GetBids(ctx sdk.Context, status types.BidStatus) []*types.Bid {
	bids := make([]*types.Bid, 0)

	k.IterateBids(ctx, func(bid *types.Bid) (stop bool) {
		if bid.Status == status {
			bids = append(bids, bid)
		}

		return false
	})

	return bids
}

// GetPendingBids gets the pending bids of the specified auction, sorted by time(asc) and price(desc)
func (k Keeper) GetPendingBids(ctx sdk.Context, auctionId uint64) []*types.Bid {
	bids := make([]*types.Bid, 0)

	k.IterateBidsByAuction(ctx, auctionId, types.BidStatus_BID_STATUS_BIDDING, func(bid *types.Bid) (stop bool) {
		bids = append(bids, bid)
		return false
	})

	if len(bids) > 0 {
		sort.SliceStable(bids, func(i, j int) bool {
			return bids[i].BidPrice > bids[j].BidPrice
		})
	}

	return bids
}

// GetAcceptedBids gets the accepted bids of the specified auction
func (k Keeper) GetAcceptedBids(ctx sdk.Context, auctionId uint64) []*types.Bid {
	bids := make([]*types.Bid, 0)

	k.IterateBidsByAuction(ctx, auctionId, types.BidStatus_BID_STATUS_ACCEPTED, func(bid *types.Bid) (stop bool) {
		bids = append(bids, bid)
		return false
	})

	return bids
}

// IterateBids iterates through all bids
func (k Keeper) IterateBids(ctx sdk.Context, cb func(bid *types.Bid) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.BidKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var bid types.Bid
		k.cdc.MustUnmarshal(iterator.Value(), &bid)

		if cb(&bid) {
			break
		}
	}
}

// IterateBidsByAuction iterates through bids by the specified auction and status
func (k Keeper) IterateBidsByAuction(ctx sdk.Context, auctionId uint64, status types.BidStatus, cb func(bid *types.Bid) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	keyPrefix := append(append(types.BidByAuctionKeyPrefix, sdk.Uint64ToBigEndian(auctionId)...), sdk.Uint64ToBigEndian(uint64(status))...)

	iterator := storetypes.KVStorePrefixIterator(store, keyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		bid := k.GetBid(ctx, sdk.BigEndianToUint64(key[1+8+8:]))

		if cb(bid) {
			break
		}
	}
}
