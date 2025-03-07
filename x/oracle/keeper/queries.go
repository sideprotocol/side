package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/oracle/types"
)

var _ types.QueryServer = Keeper{}

// Params implements types.QueryServer.
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

// ListPrices implements types.QueryServer.
func (k Keeper) ListPrices(goCtx context.Context, req *types.QueryListPricesRequest) (*types.QueryListPricesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryListPricesResponse{}, nil
}

// GetPriceBySymbol implements types.QueryServer.
func (k Keeper) GetPriceBySymbol(goCtx context.Context, req *types.QueryGetPriceBySymbolRequest) (*types.QueryGetPriceBySymbolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	price, err := k.GetPrice(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetPriceBySymbolResponse{Price: price.String()}, nil
}

// QueryBlockHeaderByHeight implements types.QueryServer.
func (k Keeper) QueryBlockHeaderByHeight(goCtx context.Context, req *types.QueryBlockHeaderByHeightRequest) (*types.QueryBlockHeaderByHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	block_header := k.GetBlockHeaderByHeight(ctx, int32(req.Height))
	if block_header == nil {
		return nil, status.Error(codes.NotFound, "block header not found")
	}

	return &types.QueryBlockHeaderByHeightResponse{BlockHeader: block_header}, nil
}

// QueryBlockHeaderByHash implements types.QueryServer.
func (k Keeper) QueryBlockHeaderByHash(goCtx context.Context, req *types.QueryBlockHeaderByHashRequest) (*types.QueryBlockHeaderByHashResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	block_header := k.GetBlockHeader(ctx, req.Hash)
	if block_header == nil {
		return nil, status.Error(codes.NotFound, "block header not found")
	}

	return &types.QueryBlockHeaderByHashResponse{BlockHeader: block_header}, nil
}

// QueryBestBlockHeader implements types.QueryServer.
func (k Keeper) QueryBestBlockHeader(goCtx context.Context, req *types.QueryBestBlockHeaderRequest) (*types.QueryBestBlockHeaderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	block_header := k.GetBestBlockHeader(ctx)
	if block_header == nil {
		return nil, status.Error(codes.NotFound, "best block header not found")
	}

	return &types.QueryBestBlockHeaderResponse{BlockHeader: block_header}, nil
}
