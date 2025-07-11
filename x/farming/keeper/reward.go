package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/farming/types"
)

// GetPendingReward gets the pending reward of the given staking
// Assume that the given staking is valid
func (k Keeper) GetPendingReward(ctx sdk.Context, stakingId uint64) sdk.Coin {
	staking := k.GetStaking(ctx, stakingId)
	totalStaking := k.GetTotalStaking(ctx, staking.PhaseId, staking.Amount.Denom)

	phase := k.GetPhase(ctx, staking.PhaseId)
	totalRewardsForAsset := phase.RewardsPerInterval.Amount.ToLegacyDec().Mul(types.GetAsset(phase, staking.Amount.Denom).RewardRatio).TruncateInt()

	rewardAmount := totalRewardsForAsset.Mul(staking.EffectiveAmount.Amount).Quo(totalStaking.EffectiveAmount.Amount)

	return sdk.NewCoin(staking.Amount.Denom, rewardAmount)
}
