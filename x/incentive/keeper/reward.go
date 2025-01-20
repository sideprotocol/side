package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/incentive/types"
)

// GetRewards gets the rewards of the given address
func (k Keeper) GetRewards(ctx sdk.Context, address string) *types.Rewards {
	store := ctx.KVStore(k.storeKey)

	var rewards types.Rewards
	bz := store.Get(types.RewardsKey(address))
	k.cdc.MustUnmarshal(bz, &rewards)

	return &rewards
}

// HasRewards returns true if the given address has received rewards, false otherwise
func (k Keeper) HasRewards(ctx sdk.Context, address string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.RewardsKey(address))
}

// SetRewards sets the given rewards
func (k Keeper) SetRewards(ctx sdk.Context, rewards *types.Rewards) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(rewards)

	store.Set(types.RewardsKey(rewards.Address), bz)
}

// AddDepositReward adds the deposit reward for the specified address by the given amount
func (k Keeper) AddDepositReward(ctx sdk.Context, address string, amount sdk.Coin) {
	rewards := k.GetRewards(ctx, address)

	if len(rewards.Address) == 0 {
		rewards.Address = address
	}

	rewards.DepositCount += 1
	rewards.DepositReward = amount.AddAmount(rewards.DepositReward.Amount)
	rewards.TotalAmount = amount.AddAmount(rewards.TotalAmount.Amount)

	k.SetRewards(ctx, rewards)
}

// AddWithdrawReward adds the withdrawal reward for the specified address by the given amount
func (k Keeper) AddWithdrawReward(ctx sdk.Context, address string, amount sdk.Coin) {
	rewards := k.GetRewards(ctx, address)

	if len(rewards.Address) == 0 {
		rewards.Address = address
	}

	rewards.WithdrawCount += 1
	rewards.WithdrawReward = amount.AddAmount(rewards.WithdrawReward.Amount)
	rewards.TotalAmount = amount.AddAmount(rewards.TotalAmount.Amount)

	k.SetRewards(ctx, rewards)
}

// GetRewardStats gets the reward statistics
func (k Keeper) GetRewardStats(ctx sdk.Context) *types.RewardStats {
	store := ctx.KVStore(k.storeKey)

	var stats types.RewardStats
	bz := store.Get(types.RewardStatsKey)
	k.cdc.MustUnmarshal(bz, &stats)

	return &stats
}

// SetRewardStats sets the reward statistics
func (k Keeper) SetRewardStats(ctx sdk.Context, rewardStats *types.RewardStats) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(rewardStats)

	store.Set(types.RewardStatsKey, bz)
}

// UpdateRewardStats updates the reward statistics
func (k Keeper) UpdateRewardStats(ctx sdk.Context, address string, reward sdk.Coin) {
	stats := k.GetRewardStats(ctx)

	if !k.HasRewards(ctx, address) {
		stats.AddressCount += 1
	}

	stats.TxCount += 1
	stats.TotalRewardAmount = reward.AddAmount(stats.TotalRewardAmount.Amount)

	k.SetRewardStats(ctx, stats)
}

// DistributeDepositReward distributes reward for deposit
func (k Keeper) DistributeDepositReward(ctx sdk.Context, address string) error {
	if !k.DepositIncentiveEnabled(ctx) {
		return types.ErrDepositIncentiveNotEnabled
	}

	rewardAmount := k.RewardPerDeposit(ctx)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(address), sdk.NewCoins(rewardAmount)); err != nil {
		return err
	}

	k.AddDepositReward(ctx, address, rewardAmount)
	k.UpdateRewardStats(ctx, address, rewardAmount)

	return nil
}

// DistributeWithdrawReward distributes reward for withdrawal
func (k Keeper) DistributeWithdrawReward(ctx sdk.Context, address string) error {
	if !k.WithdrawIncentiveEnabled(ctx) {
		return types.ErrWithdrawIncentiveNotEnabled
	}

	rewardAmount := k.RewardPerWithdraw(ctx)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(address), sdk.NewCoins(rewardAmount)); err != nil {
		return err
	}

	k.AddWithdrawReward(ctx, address, rewardAmount)
	k.UpdateRewardStats(ctx, address, rewardAmount)

	return nil
}
