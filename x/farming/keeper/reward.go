package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetPendingReward gets the pending reward of the given staking
func (k Keeper) GetPendingReward(ctx sdk.Context, stakingId uint64) sdk.Coin {
	staking := k.GetStaking(ctx, stakingId)
	totalStakings := k.GetTotalStakings(ctx, staking.Amount.Denom)

	rewardAmount := k.RewardsPerInterval(ctx).Amount.Mul(staking.EffectiveAmount.Amount).Quo(totalStakings.EffectiveAmount.Amount)

	return sdk.NewCoin(staking.Amount.Denom, rewardAmount)
}
