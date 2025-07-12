package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/farming/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) Staking(goCtx context.Context, req *types.QueryStakingRequest) (*types.QueryStakingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasStaking(ctx, req.Id) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("staking %d does not exist", req.Id))
	}

	return &types.QueryStakingResponse{Staking: k.GetStaking(ctx, req.Id)}, nil
}

func (k Keeper) Stakings(goCtx context.Context, req *types.QueryStakingsRequest) (*types.QueryStakingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryStakingsResponse{Stakings: k.GetStakingsByAddress(ctx, req.Address)}, nil
}

func (k Keeper) TotalStaking(goCtx context.Context, req *types.QueryTotalStakingRequest) (*types.QueryTotalStakingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasTotalStaking(ctx, req.Denom) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("total staking for denom %s does not exist", req.Denom))
	}

	return &types.QueryTotalStakingResponse{TotalStaking: k.GetTotalStaking(ctx, req.Denom)}, nil
}

func (k Keeper) PendingReward(goCtx context.Context, req *types.QueryPendingRewardRequest) (*types.QueryPendingRewardResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasStaking(ctx, req.Id) {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("staking %d does not exist", req.Id))
	}

	staking := k.GetStaking(ctx, req.Id)

	pendingReward := sdk.Coin{}
	if staking.Status == types.StakingStatus_STAKING_STATUS_STAKED {
		pendingReward = k.GetPendingReward(ctx, staking.Id)
	}

	return &types.QueryPendingRewardResponse{PendingReward: pendingReward.String()}, nil
}

func (k Keeper) CurrentEpoch(goCtx context.Context, req *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryCurrentEpochResponse{CurrentEpoch: k.GetCurrentEpoch(ctx)}, nil
}
