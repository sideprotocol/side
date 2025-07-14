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
	if p.EpochDuration <= 0 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid epoch duration")
	}

	if !p.RewardsPerEpoch.IsValid() || !p.RewardsPerEpoch.IsPositive() {
		return errorsmod.Wrap(ErrInvalidParams, "invalid rewards per epoch")
	}

	lockDurations := make(map[time.Duration]bool)
	for _, lockDuration := range p.LockDurations {
		if lockDurations[lockDuration] {
			return errorsmod.Wrap(ErrInvalidParams, "duplicate lock duration")
		}

		if lockDuration <= 0 {
			return errorsmod.Wrap(ErrInvalidParams, "invalid lock duration")
		}

		lockDurations[lockDuration] = true
	}

	eligibleAssets := make(map[string]bool)
	for _, asset := range p.EligibleAssets {
		if eligibleAssets[asset.Denom] {
			return errorsmod.Wrap(ErrInvalidParams, "duplicate asset denom")
		}

		if err := sdk.ValidateDenom(asset.Denom); err != nil {
			return errorsmod.Wrapf(ErrInvalidParams, "invalid asset denom: %v", err)
		}

		if asset.RewardRatio.IsNil() || asset.RewardRatio.IsZero() || asset.RewardRatio.GTE(sdkmath.LegacyOneDec()) {
			return errorsmod.Wrap(ErrInvalidParams, "invalid asset reward ratio")
		}

		eligibleAssets[asset.Denom] = true
	}

	return nil
}
