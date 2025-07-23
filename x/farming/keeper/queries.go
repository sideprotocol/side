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

	stakings, pagination, err := k.GetStakingsByStatusWithPagination(ctx, req.Status, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryStakingsResponse{Stakings: stakings, Pagination: pagination}, nil
}

func (k Keeper) StakingsByAddress(goCtx context.Context, req *types.QueryStakingsByAddressRequest) (*types.QueryStakingsByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	stakings, pagination, err := k.GetStakingsByAddressWithPagination(ctx, req.Address, req.Status, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryStakingsByAddressResponse{Stakings: stakings, Pagination: pagination}, nil
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

func (k Keeper) CurrentEpoch(goCtx context.Context, req *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryCurrentEpochResponse{CurrentEpoch: k.GetCurrentEpoch(ctx)}, nil
}

func (k Keeper) Rewards(goCtx context.Context, req *types.QueryRewardsRequest) (*types.QueryRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pendingRewards, totalRewards := k.GetRewards(ctx, req.Address)

	return &types.QueryRewardsResponse{
		PendingRewards: pendingRewards.String(),
		TotalRewards:   totalRewards.String(),
	}, nil
}

func (k Keeper) PendingReward(goCtx context.Context, req *types.QueryPendingRewardRequest) (*types.QueryPendingRewardResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasStaking(ctx, req.Id) {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("staking %d does not exist", req.Id))
	}

	if !k.HasStakingForCurrentEpoch(ctx, req.Id) {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("staking %d does not exist for the current epoch", req.Id))
	}

	staking := k.GetStaking(ctx, req.Id)

	return &types.QueryPendingRewardResponse{PendingReward: k.GetPendingReward(ctx, staking).String()}, nil
}

func (k Keeper) PendingRewardByAddress(goCtx context.Context, req *types.QueryPendingRewardByAddressRequest) (*types.QueryPendingRewardByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryPendingRewardByAddressResponse{PendingReward: k.GetPendingRewardByAddress(ctx, req.Address)}, nil
}

func (k Keeper) EstimatedReward(goCtx context.Context, req *types.QueryEstimatedRewardRequest) (*types.QueryEstimatedRewardResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	amount, err := sdk.ParseCoinNormalized(req.Amount)
	if err != nil || !amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}

	if !k.IsEligibleAsset(ctx, amount.Denom) {
		return nil, status.Error(codes.InvalidArgument, "non eligible asset")
	}

	if !k.LockDurationExists(ctx, req.LockDuration) {
		return nil, status.Error(codes.InvalidArgument, "invalid lock duration")
	}

	estimatedReward := k.GetEstimatedReward(ctx, req.Address, amount, req.LockDuration)

	return &types.QueryEstimatedRewardResponse{Reward: estimatedReward}, nil
}
