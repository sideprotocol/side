package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	
	"github.com/sideprotocol/side/x/farming/types"
)

// GetPendingReward gets the pending reward of the given staking for the current epoch
// Assume that the given staking is valid
func (k Keeper) GetPendingReward(ctx sdk.Context, staking *types.Staking) sdk.Coin {
	currentEpoch := k.GetCurrentEpoch(ctx)
	totalStaking := types.GetEpochTotalStaking(currentEpoch, staking.Amount.Denom)

	totalRewards := k.RewardsPerEpoch(ctx).Amount.ToLegacyDec().Mul(k.GetAsset(ctx, staking.Amount.Denom).RewardRatio).TruncateInt()

	rewardAmount := totalRewards.Mul(staking.EffectiveAmount.Amount).Quo(totalStaking.EffectiveAmount.Amount)

	return sdk.NewCoin(k.RewardsPerEpoch(ctx).Denom, rewardAmount)
}
