package auction

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/auction/keeper"
	"github.com/sideprotocol/side/x/auction/types"
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
		currentAuctionPrice, _, err := k.GetCurrentPrice(ctx, auction.Id)
		if err != nil {
			k.Logger(ctx).Info("Failed to get the current price", "auction id", auction.Id, "err", err)

			continue
		}

		for _, bid := range pendingBids {
			// get the remaining amount and value
			// TODO: should set the max amount for auction
			remainingAmount := auction.DepositedAsset.Amount.Int64() - auction.BiddedAmount - 10000
			remainingValue := auction.ExpectedValue.Amount.Sub(auction.BiddedValue.Amount)
			remainingAmountByPrice := remainingValue.Mul(sdkmath.NewInt(10 ^ 8)).Quo(sdkmath.NewInt(10 ^ 6)).Quo(currentAuctionPrice).Int64()

			// check if there is remaining amount and value for the auction
			if remainingAmount <= 0 || remainingValue.IsNegative() {
				auction.Status = types.AuctionStatus_AUCTION_STATUS_CLOSED
				break
			}

			// check if the bid price satisfies the current price
			if bid.BidPrice >= currentAuctionPrice.Int64() {
				biddedAmount := min(bid.BidAmount.Amount.Int64(), remainingAmount, remainingAmountByPrice)
				biddedValue := sdkmath.NewInt(bid.BidPrice).Mul(sdkmath.NewInt(biddedAmount)).Mul(sdkmath.NewInt(10 ^ 6)).Quo(sdkmath.NewInt(10 ^ 8))

				bid.BiddedAmount = sdk.NewInt64Coin(auction.DepositedAsset.Denom, biddedAmount)
				if biddedAmount == bid.BidAmount.Amount.Int64() {
					// will be updated later if partially accepted
					bid.Status = types.BidStatus_BID_STATUS_ACCEPTED
				}

				// update bid
				k.SetBid(ctx, bid)

				// accumulate the auction amount and value
				auction.BiddedAmount += biddedAmount
				auction.BiddedValue = sdk.NewCoin(auction.ExpectedValue.Denom, auction.BiddedValue.Amount.Add(biddedValue))
			}
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
			refundAmount := bid.BidAmount.SubAmount(bid.BiddedAmount.Amount)
			refundAsset := sdk.NewCoin(auction.ExpectedValue.Denom, sdkmath.NewInt(bid.BidPrice).Mul(refundAmount.Amount).Mul(sdkmath.NewInt(10^6)).Quo(sdkmath.NewInt(10^8)))

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

		// slash borrower for liquidation penalty
		currentPrice := k.GetPrice(ctx, "BTC-USD")
		if currentPrice.IsZero() {
			k.Logger(ctx).Info("Failed to get the current price", "block height", ctx.BlockHeight())
		} else {
			remainingAmount := auction.DepositedAsset.Amount.Int64() - auction.BiddedAmount - 10000
			slashedValue := currentPrice.Mul(sdkmath.NewInt(remainingAmount)).Mul(sdkmath.NewInt(10 ^ 6)).Mul(sdkmath.NewInt(int64(k.GetParams(ctx).LiquidationBonus))).Quo(sdkmath.NewInt(10 ^ 8)).Quo(sdkmath.NewInt(1000))

			slashedAsset := sdk.NewCoin(auction.ExpectedValue.Denom, slashedValue)
			if err := k.BankKeeper().SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(auction.Borrower), types.ModuleName, sdk.NewCoins(slashedAsset)); err != nil {
				k.Logger(ctx).Info("Failed to slash borrower", "auction id", auction.Id, "borrower", auction.Borrower, "amount", slashedAsset, "err", err)
			}
		}

		// handle bidded asset(repay the lending pool)
		biddedAsset := auction.BiddedValue
		if err := k.BiddedAssetHandler()(ctx, auction.LoanId, types.ModuleName, biddedAsset); err != nil {
			k.Logger(ctx).Info("Failed to call BiddedAssetHandler", "auction id", auction.Id, "amount", biddedAsset, "err", err)

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
