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

	// ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryGetPriceBySymbolResponse{}, nil
}
