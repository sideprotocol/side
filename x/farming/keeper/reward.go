package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/farming/types"
)

// GetPendingReward gets the pending reward of the given staking for the current epoch
// Assume that the given staking is valid
func (k Keeper) GetPendingReward(ctx sdk.Context, staking *types.Staking) sdk.Coin {
	currentEpoch := k.GetCurrentEpoch(ctx)

	totalStaking := types.GetEpochTotalStaking(currentEpoch, staking.Amount.Denom)
	asset := types.GetAsset(k.EligibleAssets(ctx), staking.Amount.Denom)

	totalRewards := k.RewardPerEpoch(ctx).Amount.ToLegacyDec().Mul(asset.RewardRatio).TruncateInt()

	rewardAmount := totalRewards.Mul(staking.EffectiveAmount.Amount).Quo(totalStaking.EffectiveAmount.Amount)

	return sdk.NewCoin(k.RewardPerEpoch(ctx).Denom, rewardAmount)
}

// GetPendingRewardByAddress gets the pending reward of the given address for the current epoch
func (k Keeper) GetPendingRewardByAddress(ctx sdk.Context, address string) *types.AccountRewardPerEpoch {
	currentEpoch := k.GetCurrentEpoch(ctx)

	totalStakings := []types.TotalStaking{}

	k.IterateCurrentEpochStakingQueueByAddress(ctx, address, func(staking *types.Staking) (stop bool) {
		totalStakings = types.UpdateAccountTotalStakings(totalStakings, staking)
		return false
	})

	if len(totalStakings) == 0 {
		return nil
	}

	return types.GetAccountRewardPerEpoch(address, totalStakings, currentEpoch, k.RewardPerEpoch(ctx), k.EligibleAssets(ctx))
}

// GetRewards gets the reward stats of the given address
func (k Keeper) GetRewards(ctx sdk.Context, address string) (sdk.Coin, sdk.Coin) {
	stakings := k.GetStakingsByAddress(ctx, address)

	pendingRewards := sdk.NewCoin(k.RewardPerEpoch(ctx).Denom, sdkmath.ZeroInt())
	totalRewards := pendingRewards

	for _, staking := range stakings {
		pendingRewards = pendingRewards.Add(staking.PendingRewards)
		totalRewards = totalRewards.Add(staking.TotalRewards)
	}

	return pendingRewards, totalRewards
}
