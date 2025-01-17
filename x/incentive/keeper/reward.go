package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/incentive/types"
)

// SetReward sets the given reward
func (k Keeper) SetReward(ctx sdk.Context, reward *types.Reward) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(reward)

	store.Set(types.RewardKey(reward.Address), bz)
}

// GetReward gets the reward of the given address
func (k Keeper) GetReward(ctx sdk.Context, address string) *types.Reward {
	store := ctx.KVStore(k.storeKey)

	var reward types.Reward
	bz := store.Get(types.RewardKey(address))
	k.cdc.MustUnmarshal(bz, &reward)

	return &reward
}

// SetTotalRewards sets the total rewards
func (k Keeper) SetTotalRewards(ctx sdk.Context, totalRewards sdk.Coin) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&totalRewards)

	store.Set(types.TotalRewardsKey, bz)
}

// GetTotalRewards gets the total rewards
func (k Keeper) GetTotalRewards(ctx sdk.Context) sdk.Coin {
	store := ctx.KVStore(k.storeKey)

	var rewards sdk.Coin
	bz := store.Get(types.TotalRewardsKey)
	k.cdc.MustUnmarshal(bz, &rewards)

	return rewards
}

// UpdateTotalRewards updates the total rewards by delta
func (k Keeper) UpdateTotalRewards(ctx sdk.Context, delta sdk.Coin) {
	currentTotalRewards := k.GetTotalRewards(ctx)
	newTotalRewards := delta.AddAmount(currentTotalRewards.Amount)

	k.SetTotalRewards(ctx, newTotalRewards)
}

// DistributeDepositReward distributes reward for deposit
func (k Keeper) DistributeDepositReward(ctx sdk.Context, address string) error {
	if !k.IncentiveEnabled(ctx) {
		return nil
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(address), sdk.NewCoins(k.RewardPerDeposit(ctx))); err != nil {
		// ignore error
		return nil
	}

	reward := k.GetReward(ctx, address)

	reward.Address = address
	reward.DepositCount += 1
	reward.TotalAmount = k.RewardPerDeposit(ctx).AddAmount(reward.TotalAmount.Amount)

	k.SetReward(ctx, reward)
	k.UpdateTotalRewards(ctx, k.RewardPerDeposit(ctx))

	return nil
}

// DistributeWithdrawReward distributes reward for withdrawal
func (k Keeper) DistributeWithdrawReward(ctx sdk.Context, address string) error {
	if !k.IncentiveEnabled(ctx) {
		return nil
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(address), sdk.NewCoins(k.RewardPerWithdraw(ctx))); err != nil {
		// ignore error
		return nil
	}

	reward := k.GetReward(ctx, address)

	reward.Address = address
	reward.WithdrawCount += 1
	reward.TotalAmount = k.RewardPerWithdraw(ctx).AddAmount(reward.TotalAmount.Amount)

	k.SetReward(ctx, reward)
	k.UpdateTotalRewards(ctx, k.RewardPerWithdraw(ctx))

	return nil
}
