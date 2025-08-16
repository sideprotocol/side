package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// duration for 1 day
	Day = 24 * time.Hour

	// default epoch duration
	DefaultEpochDuration = 7 * Day

	// default reward per epoch
	DefaultRewardPerEpoch = sdk.NewCoin("uside", sdkmath.NewIntWithDecimal(0, 6)) // 0 SIDE

	// default lock durations
	DefaultLockDurations = []time.Duration{
		7 * Day, 30 * Day, 60 * Day, 90 * Day,
		120 * Day, 180 * Day, 365 * Day,
	}
)

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		Enabled:        false,
		EpochDuration:  DefaultEpochDuration,
		RewardPerEpoch: DefaultRewardPerEpoch,
		LockDurations:  DefaultLockDurations,
		EligibleAssets: nil,
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateEpochDuration(p); err != nil {
		return err
	}

	if err := validateRewardPerEpoch(p); err != nil {
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

// ValidateParamsUpdate validates the given params update
func ValidateParamsUpdate(params Params, newParams Params) error {
	if newParams.RewardPerEpoch.Denom != params.RewardPerEpoch.Denom {
		return errorsmod.Wrap(ErrInvalidParams, "reward denom cannot be updated")
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

	return nil
}

// validateRewardPerEpoch validates the given reward per epoch
func validateRewardPerEpoch(p Params) error {
	if !p.RewardPerEpoch.IsValid() {
		return errorsmod.Wrap(ErrInvalidParams, "invalid reward per epoch")
	}

	if p.Enabled && !p.RewardPerEpoch.IsPositive() {
		return errorsmod.Wrap(ErrInvalidParams, "reward per epoch must be positive when farming enabled")
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

		if err := validateRewardRatio(asset.RewardRatio); err != nil {
			return err
		}

		if !asset.MinStakingAmount.IsPositive() {
			return errorsmod.Wrapf(ErrInvalidParams, "min staking amount must be positive")
		}

		totalRewardRatio = totalRewardRatio.Add(asset.RewardRatio)
		eligibleAssets[asset.Denom] = true
	}

	if totalRewardRatio.GT(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidParams, "total asset reward ratio cannot be greater than 1")
	}

	return nil
}

// validateRewardRatio validates the given reward ratio
func validateRewardRatio(rewardRatio sdkmath.LegacyDec) error {
	if rewardRatio.IsNegative() {
		return errorsmod.Wrap(ErrInvalidParams, "asset reward ratio cannot be negative")
	}

	if rewardRatio.GT(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidParams, "asset reward ratio cannot be greater than 1")
	}

	return nil
}
