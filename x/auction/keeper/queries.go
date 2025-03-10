package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/auction/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) Auction(goCtx context.Context, req *types.QueryAuctionRequest) (*types.QueryAuctionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasAuction(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "auction does not exist")
	}

	return &types.QueryAuctionResponse{Auction: k.GetAuction(ctx, req.Id)}, nil
}

func (k Keeper) Auctions(goCtx context.Context, req *types.QueryAuctionsRequest) (*types.QueryAuctionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryAuctionsResponse{Auctions: k.GetAuctions(ctx, req.Status)}, nil
}

func (k Keeper) Bid(goCtx context.Context, req *types.QueryBidRequest) (*types.QueryBidResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasBid(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "bid does not exist")
	}

	return &types.QueryBidResponse{Bid: k.GetBid(ctx, req.Id)}, nil
}

func (k Keeper) Bids(goCtx context.Context, req *types.QueryBidsRequest) (*types.QueryBidsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var bids []*types.Bid

	if req.AuctionId == 0 {
		if req.Status == types.BidStatus_BID_STATUS_UNSPECIFIED {
			bids = k.GetAllBids(ctx)
		} else {
			bids = k.GetBids(ctx, req.Status)
		}
	} else {
		bids = k.GetBidsByAuction(ctx, req.AuctionId, req.Status)
	}

	return &types.QueryBidsResponse{Bids: bids}, nil
}

func (k Keeper) AuctionPrice(goCtx context.Context, req *types.QueryAuctionPriceRequest) (*types.QueryAuctionPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasAuction(ctx, req.AuctionId) {
		return nil, status.Error(codes.InvalidArgument, "auction does not exist")
	}

	if k.GetAuction(ctx, req.AuctionId).Status != types.AuctionStatus_AUCTION_STATUS_OPEN {
		return nil, status.Error(codes.InvalidArgument, "auction not open")
	}

	price, err := k.GetCurrentPrice(ctx, req.AuctionId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAuctionPriceResponse{Price: price.String()}, nil
}
