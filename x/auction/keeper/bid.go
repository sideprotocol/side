package keeper

import (
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
	if auction.Status == types.AuctionStatus_AuctionClose {
		return nil, types.ErrAuctionClosed
	}

	if amount.Amount.Uint64() < k.GetParams(ctx).MinBidAmount {
		return nil, errorsmod.Wrap(types.ErrInvalidBid, "amount can not be less than the minimum allowed amount")
	}

	bid := &types.Bid{
		Id:        k.IncrementBidId(ctx),
		Bidder:    sender,
		AuctionId: auctionId,
		BidPrice:  price,
		BidAmount: amount,
		Status:    types.BidStatus_Bidding,
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

	if bid.Status != types.BidStatus_Bidding {
		return types.ErrInvalidBidStatus
	}

	bid.Status = types.BidStatus_Rejected
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

	bz := k.cdc.MustMarshal(bid)
	store.Set(types.BidKey(bid.Id), bz)
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

// IterateBids iterates through all bids
func (k Keeper) IterateBids(ctx sdk.Context, cb func(req *types.Bid) (stop bool)) {
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
