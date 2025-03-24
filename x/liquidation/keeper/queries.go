package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) Liquidation(goCtx context.Context, req *types.QueryLiquidationRequest) (*types.QueryLiquidationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLiquidation(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "liquidation does not exist")
	}

	return &types.QueryLiquidationResponse{Liquidation: k.GetLiquidation(ctx, req.Id)}, nil
}

func (k Keeper) Liquidations(goCtx context.Context, req *types.QueryLiquidationsRequest) (*types.QueryLiquidationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var liquidations []*types.Liquidation

	if req.Status == types.LiquidationStatus_LIQUIDATION_STATUS_UNSPECIFIED {
		liquidations = k.GetAllLiquidations(ctx)
	} else {
		liquidations = k.GetLiquidations(ctx, req.Status)
	}

	return &types.QueryLiquidationsResponse{Liquidations: liquidations}, nil
}

func (k Keeper) LiquidationRecord(goCtx context.Context, req *types.QueryLiquidationRecordRequest) (*types.QueryLiquidationRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLiquidationRecord(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, "liquidation record does not exist")
	}

	return &types.QueryLiquidationRecordResponse{LiquidationRecord: k.GetLiquidationRecord(ctx, req.Id)}, nil
}

func (k Keeper) LiquidationRecords(goCtx context.Context, req *types.QueryLiquidationRecordsRequest) (*types.QueryLiquidationRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var records []*types.LiquidationRecord

	if req.LiquidationId == 0 {
		records = k.GetAllLiquidationRecords(ctx)
	} else {
		records = k.GetLiquidationRecords(ctx, req.LiquidationId)
	}

	return &types.QueryLiquidationRecordsResponse{LiquidationRecords: records}, nil
}
