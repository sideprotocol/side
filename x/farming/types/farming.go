package types

import (
	"slices"
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
	lockDurationInDays := lockDuration / (24 * time.Hour)

	return sdkmath.LegacyNewDec(int64(lockDurationInDays)).QuoInt64(365).Mul(LockMultiplierFactor).Add(sdkmath.LegacyOneDec())
}

// GetEffectiveAmount gets the effective staked amount according to the given amount and lock multiplier
// Formula: effective amount = amount * multiplier
func GetEffectiveAmount(amount sdk.Coin, lockMultiplier sdkmath.LegacyDec) sdk.Coin {
	effectiveAmount := amount.Amount.ToLegacyDec().Mul(lockMultiplier).TruncateInt()

	return sdk.NewCoin(amount.Denom, effectiveAmount)
}

// LockDurationExists returns true if the given lock duration exists for the specified phase, false otherwise
func LockDurationExists(phase *Phase, lockDuration time.Duration) bool {
	return slices.Contains(phase.LockDurations, lockDuration)
}

// IsAllowedAsset returns true if the given asset is allowed for the specified phase, false otherwise
func IsAllowedAsset(phase *Phase, denom string) bool {
	for _, asset := range phase.AllowedAssets {
		if asset.Denom == denom {
			return true
		}
	}

	return false
}

// GetAsset gets the allowed asset for the given phase and denom
func GetAsset(phase *Phase, denom string) *Asset {
	for _, asset := range phase.AllowedAssets {
		if asset.Denom == denom {
			return &asset
		}
	}

	return nil
}
