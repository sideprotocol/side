package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetPendingReward gets the pending reward of the given staking for the current epoch
// Assume that the given staking is valid
func (k Keeper) GetPendingReward(ctx sdk.Context, stakingId uint64) sdk.Coin {
	staking := k.GetStaking(ctx, stakingId)
	totalStaking := k.GetTotalStaking(ctx, staking.Amount.Denom)

	totalRewards := k.RewardsPerEpoch(ctx).Amount.ToLegacyDec().Mul(k.GetAsset(ctx, staking.Amount.Denom).RewardRatio).TruncateInt()

	rewardAmount := totalRewards.Mul(staking.EffectiveAmount.Amount).Quo(totalStaking.EffectiveAmount.Amount)

	return sdk.NewCoin(k.RewardsPerEpoch(ctx).Denom, rewardAmount)
}
