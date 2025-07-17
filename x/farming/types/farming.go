package types

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// Lock multiplier factor for staking
	LockMultiplierFactor = sdkmath.LegacyMustNewDecFromStr("2.5")
)

// GetLockMultiplier gets the lock multiplier according to the given lock duration
// Formula: 1 + (lockDurationInDays / 365) * 2.5
func GetLockMultiplier(lockDuration time.Duration) sdkmath.LegacyDec {
	lockDurationInDays := GetLockDurationInDays(lockDuration)

	return sdkmath.LegacyNewDec(int64(lockDurationInDays)).QuoInt64(365).Mul(LockMultiplierFactor).Add(sdkmath.LegacyOneDec())
}

// GetEffectiveAmount gets the effective staked amount according to the given amount and lock multiplier
// Formula: effective amount = amount * multiplier
func GetEffectiveAmount(amount sdk.Coin, lockMultiplier sdkmath.LegacyDec) sdk.Coin {
	effectiveAmount := amount.Amount.ToLegacyDec().Mul(lockMultiplier).TruncateInt()

	return sdk.NewCoin(amount.Denom, effectiveAmount)
}

// GetLockDurationInDays gets days for the given lock duration
func GetLockDurationInDays(lockDuration time.Duration) time.Duration {
	return lockDuration / (24 * time.Hour)
}

// GetEpochReward calculates the reward of the given staking for the specified epoch
// Assume that the given params are valid
// Formula: rewardPerEpoch * assetRewardRatio * effectiveAmount / totalEffectiveAmount
func GetEpochReward(ctx sdk.Context, staking *Staking, epoch *Epoch, rewardPerEpoch sdk.Coin, assetRewardRatio sdkmath.LegacyDec) sdk.Coin {
	totalStaking := GetEpochTotalStaking(epoch, staking.Amount.Denom)

	totalRewards := rewardPerEpoch.Amount.ToLegacyDec().Mul(assetRewardRatio).TruncateInt()

	rewardAmount := totalRewards.Mul(staking.EffectiveAmount.Amount).Quo(totalStaking.EffectiveAmount.Amount)

	return sdk.NewCoin(rewardPerEpoch.Denom, rewardAmount)
}

// GetAsset gets the asset by the given denom
func GetAsset(assets []Asset, denom string) Asset {
	for _, asset := range assets {
		if asset.Denom == denom {
			return asset
		}
	}

	return Asset{}
}

// GetEpochTotalStaking gets the total staking for the specified epoch by the given denom
func GetEpochTotalStaking(epoch *Epoch, denom string) *TotalStaking {
	for _, totalStaking := range epoch.TotalStakings {
		if totalStaking.Denom == denom {
			return &totalStaking
		}
	}

	return nil
}

// UpdateEpochTotalStakings updates the total stakings for the specified epoch by the given staking
func UpdateEpochTotalStakings(epoch *Epoch, staking *Staking) {
	for i, totalStaking := range epoch.TotalStakings {
		if totalStaking.Denom == staking.Amount.Denom {
			// update total staking if existing
			epoch.TotalStakings[i].Amount = totalStaking.Amount.Add(staking.Amount)
			epoch.TotalStakings[i].EffectiveAmount = totalStaking.EffectiveAmount.Add(staking.EffectiveAmount)

			return
		}
	}

	// add new total staking if not found
	epoch.TotalStakings = append(epoch.TotalStakings, TotalStaking{
		Denom:           staking.Amount.Denom,
		Amount:          staking.Amount,
		EffectiveAmount: staking.EffectiveAmount,
	})
}

// UpdateAccountTotalStakings updates the account total stakings by the given staking
func UpdateAccountTotalStakings(totalStakings []TotalStaking, staking *Staking) []TotalStaking {
	for i, totalStaking := range totalStakings {
		if totalStaking.Denom == staking.Amount.Denom {
			// update total staking if existing
			totalStakings[i].Amount = totalStaking.Amount.Add(staking.Amount)
			totalStakings[i].EffectiveAmount = totalStaking.EffectiveAmount.Add(staking.EffectiveAmount)

			return totalStakings
		}
	}

	// add new total staking if not found
	return append(totalStakings, TotalStaking{
		Denom:           staking.Amount.Denom,
		Amount:          staking.Amount,
		EffectiveAmount: staking.EffectiveAmount,
	})
}

// GetAccountRewardPerEpoch gets the account reward for the given epoch
func GetAccountRewardPerEpoch(address string, accountTotalStakings []TotalStaking, epoch *Epoch, rewardPerEpoch sdk.Coin, assets []Asset) *AccountRewardPerEpoch {
	accountRewardPerEpoch := &AccountRewardPerEpoch{
		Address:    address,
		Stakings:   accountTotalStakings,
		Shares:     []sdkmath.LegacyDec{},
		TotalShare: sdkmath.LegacyZeroDec(),
		Reward:     sdk.NewCoin(rewardPerEpoch.Denom, sdkmath.ZeroInt()),
	}

	for _, totalStaking := range accountTotalStakings {
		epochTotalStaking := GetEpochTotalStaking(epoch, totalStaking.Denom)
		asset := GetAsset(assets, totalStaking.Denom)

		share := totalStaking.EffectiveAmount.Amount.ToLegacyDec().QuoInt(epochTotalStaking.EffectiveAmount.Amount)
		rewardAmount := rewardPerEpoch.Amount.ToLegacyDec().Mul(asset.RewardRatio).Mul(share).TruncateInt()

		accountRewardPerEpoch.Shares = append(accountRewardPerEpoch.Shares, share)
		accountRewardPerEpoch.TotalShare = accountRewardPerEpoch.TotalShare.Add(share.Mul(asset.RewardRatio))

		accountRewardPerEpoch.Reward = accountRewardPerEpoch.Reward.AddAmount(rewardAmount)
	}

	return accountRewardPerEpoch
}
