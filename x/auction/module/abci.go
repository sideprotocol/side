package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/auction/keeper"
	"github.com/sideprotocol/side/x/auction/types"
	lendingtypes "github.com/sideprotocol/side/x/lending/types"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleAuctions(ctx, k)
	handleBids(ctx, k)
}

// handleAuction performs the auction handling
func handleAuctions(ctx sdk.Context, k keeper.Keeper) {
	auctions := k.GetAuctions(ctx, types.AuctionStatus_AuctionClose)

	for _, auction := range auctions {
		k.IterateEscrowAssets(ctx, auction.Id, func(auctoinId, bidId uint64, asset sdk.Coin) (stop bool) {
			if err := k.BankKeeper().SendCoinsFromModuleToModule(ctx, types.ModuleName, lendingtypes.ModuleName, sdk.NewCoins(asset)); err != nil {
				k.Logger(ctx).Info("Failed to transfer asset to lending module", "auction id", auctoinId, "err", err)

				return false
			}

			k.RemoveEscrowAsset(ctx, auctoinId, bidId)

			return false
		})
	}
}

// handleBids performs the bid handling
func handleBids(ctx sdk.Context, k keeper.Keeper) {
	bids := k.GetBids(ctx, types.BidStatus_Bidding)

	for _, bid := range bids {
		auction := k.GetAuction(ctx, bid.AuctionId)
		if auction.Status != types.AuctionStatus_AuctionOpen {
			continue
		}

		price, err := k.GetCurrentPrice(ctx, bid.AuctionId)
		if err != nil {
			k.Logger(ctx).Info("Failed to get the current price", "auction id", bid.AuctionId, "err", err)

			continue
		}

		if bid.BidPrice >= price.Int64() {
			totalBidPrice := bid.BidAmount.Amount.Int64() * bid.BidPrice
			totalBidAsset := sdk.NewInt64Coin("uusdc", totalBidPrice)

			if err := k.BankKeeper().SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(bid.Bidder), types.ModuleName, sdk.NewCoins(totalBidAsset)); err != nil {
				k.Logger(ctx).Info("Failed to transfer coins for bid", "bid id", bid.Id, "err", err)

				continue
			}

			k.SetEscrowAsset(ctx, bid.AuctionId, bid.Id, totalBidAsset)

			bid.Status = types.BidStatus_Accepted
		}
	}
}
