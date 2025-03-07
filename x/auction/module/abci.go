package auction

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/auction/keeper"
	"github.com/sideprotocol/side/x/auction/types"
	lendingtypes "github.com/sideprotocol/side/x/lending/types"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handlePendingAuctions(ctx, k)
	handleCompletedAuctions(ctx, k)
}

// handlePendingAuctions handles the pending auctions
func handlePendingAuctions(ctx sdk.Context, k keeper.Keeper) {
	// get pending auctions
	pendingAuctions := k.GetAuctions(ctx, types.AuctionStatus_AUCTION_STATUS_OPEN)

	for _, auction := range pendingAuctions {
		// get pending bids
		pendingBids := k.GetPendingBids(ctx, auction.Id)
		if len(pendingBids) == 0 {
			continue
		}

		// get the current auction price
		currentAuctionPrice, err := k.GetCurrentPrice(ctx, auction.Id)
		if err != nil {
			k.Logger(ctx).Info("Failed to get the current price", "auction id", auction.Id, "err", err)

			continue
		}

		for _, bid := range pendingBids {
			if bid.BidPrice >= currentAuctionPrice.Int64() {
				bidValue := bid.BidAmount.Amount.Int64() * bid.BidPrice
				auction.BiddedValue += bidValue

				bid.BiddedAmount = bid.BidAmount
				bid.Status = types.BidStatus_BID_STATUS_ACCEPTED

				// update bid
				k.SetBid(ctx, bid)
			}
		}

		// close auction if bidded value >= expected value
		if auction.BiddedValue >= auction.ExpectedValue {
			auction.Status = types.AuctionStatus_AUCTION_STATUS_CLOSED
		}

		// update auction
		k.SetAuction(ctx, auction)
	}
}

// handleCompletedAuctions handles the completed auctions
func handleCompletedAuctions(ctx sdk.Context, k keeper.Keeper) {
	// get completed auctions
	completedAuctions := k.GetAuctions(ctx, types.AuctionStatus_AUCTION_STATUS_CLOSED)

	for _, auction := range completedAuctions {
		// get pending bids
		pendingBids := k.GetPendingBids(ctx, auction.Id)

		// refund
		for _, bid := range pendingBids {
			// must be positive result here
			refundAmount := bid.BidAmount.Sub(bid.BiddedAmount)
			refundAsset := sdk.NewInt64Coin("uusdc", bid.BidPrice*refundAmount.Amount.Int64())

			if err := k.BankKeeper().SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(bid.Bidder), sdk.NewCoins(refundAsset)); err != nil {
				k.Logger(ctx).Info("Failed to refund", "auction id", auction.Id, "bid id", bid.Id, "amount", refundAsset, "err", err)

				continue
			}

			// possibly partially accepted in case the bid is the last accepted one
			if bid.BiddedAmount.IsPositive() {
				bid.Status = types.BidStatus_BID_STATUS_ACCEPTED
			} else {
				bid.Status = types.BidStatus_BID_STATUS_REJECTED
			}

			// update bid
			k.SetBid(ctx, bid)
		}

		// transfer bidded asset to the lending pool
		biddedAsset := sdk.NewInt64Coin("uusdc", auction.BiddedValue)
		if err := k.BankKeeper().SendCoinsFromModuleToModule(ctx, types.ModuleName, lendingtypes.ModuleName, sdk.NewCoins(biddedAsset)); err != nil {
			k.Logger(ctx).Info("Failed to transfer bidded asset to lending module", "auction id", auction.Id, "amount", biddedAsset, "err", err)

			continue
		}

		// build payment tx
		paymentTx, txHash, sigHashes, err := types.BuildPaymentTransaction(auction, k.GetAcceptedBids(ctx, auction.Id), 10)
		if err != nil {
			k.Logger(ctx).Info("Failed to build payment transaction", "auction id", auction.Id, "err", err)

			continue
		}

		// emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSignPaymentTransaction,
				sdk.NewAttribute(types.AttributeKeyAuctionId, fmt.Sprintf("%d", auction.Id)),
				sdk.NewAttribute(types.AttributeKeyAgencyPubKey, auction.Agency),
				sdk.NewAttribute(types.AttributeKeySigHashes, strings.Join(sigHashes, types.AttributeValueSeparator)),
			),
		)

		auction.PaymentTx = paymentTx
		auction.PaymentTxId = txHash.String()
		auction.Status = types.AuctionStatus_AUCTION_STATUS_SETTLED

		// update auction
		k.SetAuction(ctx, auction)
	}
}
