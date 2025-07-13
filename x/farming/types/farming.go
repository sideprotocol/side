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

// GetEpochTotalStaking gets the total staking for the specified epoch by the given denom
func GetEpochTotalStaking(epoch *Epoch, denom string) *TotalStaking {
	for _, totalStaking := range epoch.TotalStakings {
		if totalStaking.Denom == denom {
			return &totalStaking
		}
	}

	return nil
}

// UpdateEpochTotalStaking updates the total staking for the specified epoch by the given staking
func UpdateEpochTotalStaking(epoch *Epoch, staking *Staking) {
	for i, totalStaking := range epoch.TotalStakings {
		if totalStaking.Denom == staking.Amount.Denom {
			// update total staking if existing
			totalStaking.Amount = totalStaking.Amount.Add(staking.Amount)
			totalStaking.EffectiveAmount = totalStaking.EffectiveAmount.Add(staking.EffectiveAmount)

			epoch.TotalStakings[i] = totalStaking
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
