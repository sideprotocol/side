package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/incentive/types"
)

// GetReward gets the reward of the given address
func (k Keeper) GetReward(ctx sdk.Context, address string) *types.Reward {
	store := ctx.KVStore(k.storeKey)

	var reward types.Reward
	bz := store.Get(types.RewardKey(address))
	k.cdc.MustUnmarshal(bz, &reward)

	return &reward
}

// SetReward sets the given reward
func (k Keeper) SetReward(ctx sdk.Context, reward *types.Reward) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(reward)

	store.Set(types.RewardKey(reward.Address), bz)
}

// AddDepositReward adds the deposit reward for the specified address by the given amount
func (k Keeper) AddDepositReward(ctx sdk.Context, address string, amount sdk.Coin) {
	reward := k.GetReward(ctx, address)

	reward.Address = address
	reward.DepositCount += 1
	reward.TotalAmount = amount.AddAmount(reward.TotalAmount.Amount)

	k.SetReward(ctx, reward)
}

// AddWithdrawReward adds the withdrawal reward for the specified address by the given amount
func (k Keeper) AddWithdrawReward(ctx sdk.Context, address string, amount sdk.Coin) {
	reward := k.GetReward(ctx, address)

	reward.Address = address
	reward.WithdrawCount += 1
	reward.TotalAmount = amount.AddAmount(reward.TotalAmount.Amount)

	k.SetReward(ctx, reward)
}

// GetTotalRewards gets the total rewards
func (k Keeper) GetTotalRewards(ctx sdk.Context) sdk.Coin {
	store := ctx.KVStore(k.storeKey)

	var totalRewards sdk.Coin
	bz := store.Get(types.TotalRewardsKey)
	k.cdc.MustUnmarshal(bz, &totalRewards)

	return totalRewards
}

// SetTotalRewards sets the total rewards
func (k Keeper) SetTotalRewards(ctx sdk.Context, totalRewards sdk.Coin) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&totalRewards)

	store.Set(types.TotalRewardsKey, bz)
}

// UpdateTotalRewards updates the total rewards by delta
func (k Keeper) UpdateTotalRewards(ctx sdk.Context, delta sdk.Coin) {
	currentTotalRewards := k.GetTotalRewards(ctx)
	newTotalRewards := delta.AddAmount(currentTotalRewards.Amount)

	k.SetTotalRewards(ctx, newTotalRewards)
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
	k.UpdateTotalRewards(ctx, rewardAmount)

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
	k.UpdateTotalRewards(ctx, rewardAmount)

	return nil
}
