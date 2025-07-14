package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// default epoch duration
	DefaultEpochDuration = 7 * 24 * time.Hour // 1W

	// default rewards per epoch
	DefaultRewardsPerEpoch = sdk.NewCoin("uside", sdkmath.NewIntWithDecimal(1000000, 6)) // 100000 SIDE

	// default lock durations
	DefaultLockDurations = []time.Duration{30 * 24 * time.Hour} // 30 days
)

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		Enabled:         false,
		EpochDuration:   DefaultEpochDuration,
		RewardsPerEpoch: DefaultRewardsPerEpoch,
		LockDurations:   DefaultLockDurations,
		EligibleAssets:  nil,
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateEpochDuration(p); err != nil {
		return err
	}

	if err := validateRewardsPerEpoch(p); err != nil {
		return err
	}

	if err := validateLockDurations(p); err != nil {
		return err
	}

	if err := validateEligibleAssets(p); err != nil {
		return err
	}

	return nil
}

// validateEpochDuration validates the given epoch duration
func validateEpochDuration(p Params) error {
	if p.EpochDuration < 0 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid epoch duration")
	}

	if p.Enabled && p.EpochDuration == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "epoch duration must be greater than 0 when farming enabled")
	}
}

// validateRewardsPerEpoch validates the given rewards per epoch
func validateRewardsPerEpoch(p Params) error {
	if !p.RewardsPerEpoch.IsValid() {
		return errorsmod.Wrap(ErrInvalidParams, "invalid rewards per epoch")
	}

	if p.Enabled && !p.RewardsPerEpoch.IsPositive() {
		return errorsmod.Wrap(ErrInvalidParams, "rewards per epoch must be positive when farming enabled")
	}

	return nil
}

// validateLockDurations validates the given lock durations
func validateLockDurations(p Params) error {
	if p.Enabled && len(p.LockDurations) == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "lock durations cannot be empty when farming enabled")
	}

	lockDurations := make(map[time.Duration]bool)
	for _, lockDuration := range p.LockDurations {
		if lockDurations[lockDuration] {
			return errorsmod.Wrap(ErrInvalidParams, "duplicate lock duration")
		}

		if lockDuration < p.EpochDuration {
			return errorsmod.Wrap(ErrInvalidParams, "lock duration cannot be less than epoch duration")
		}

		lockDurations[lockDuration] = true
	}

	return nil
}

// validateEligibleAssets validates the given eligible assets
func validateEligibleAssets(p Params) error {
	if p.Enabled && len(p.EligibleAssets) == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "eligible assets cannot be empty when farming enabled")
	}

	eligibleAssets := make(map[string]bool)
	totalRewardRatio := sdkmath.LegacyZeroDec()

	for _, asset := range p.EligibleAssets {
		if eligibleAssets[asset.Denom] {
			return errorsmod.Wrap(ErrInvalidParams, "duplicate asset denom")
		}

		if err := sdk.ValidateDenom(asset.Denom); err != nil {
			return errorsmod.Wrapf(ErrInvalidParams, "invalid asset denom: %v", err)
		}

		if asset.RewardRatio.IsNegative() {
			return errorsmod.Wrap(ErrInvalidParams, "asset reward ratio cannot be negative")
		}

		if asset.RewardRatio.GT(sdkmath.LegacyOneDec()) {
			return errorsmod.Wrap(ErrInvalidParams, "asset reward ratio cannot be greater than 1")
		}

		totalRewardRatio = totalRewardRatio.Add(asset.RewardRatio)
		eligibleAssets[asset.Denom] = true
	}

	if totalRewardRatio.GT(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidParams, "total asset reward ratio cannot be greater than 1")
	}

	return nil
}
