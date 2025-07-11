package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreatePhase{}

// ValidateBasic performs basic message validation.
func (m *MsgCreatePhase) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if m.StartTime.IsZero() {
		return errorsmod.Wrap(ErrInvalidPhaseParams, "invalid start time")
	}

	if m.Duration <= 0 {
		return errorsmod.Wrap(ErrInvalidPhaseParams, "invalid duration")
	}

	if m.DistributionInterval <= 0 {
		return errorsmod.Wrap(ErrInvalidPhaseParams, "invalid distribution interval")
	}

	if !m.RewardsPerInterval.IsValid() || !m.RewardsPerInterval.IsPositive() {
		return errorsmod.Wrap(ErrInvalidPhaseParams, "invalid rewards per interval")
	}

	lockDurations := make(map[time.Duration]bool)
	for _, lockDuration := range m.LockDurations {
		if lockDurations[lockDuration] {
			return errorsmod.Wrap(ErrInvalidPhaseParams, "duplicate lock duration")
		}

		if lockDuration <= 0 {
			return errorsmod.Wrap(ErrInvalidPhaseParams, "invalid lock duration")
		}

		if lockDuration > m.Duration {
			return errorsmod.Wrap(ErrInvalidPhaseParams, "lock duration cannot be greater than the phase duration")
		}

		lockDurations[lockDuration] = true
	}

	allowedAssets := make(map[string]bool)
	for _, asset := range m.AllowedAssets {
		if allowedAssets[asset.Denom] {
			return errorsmod.Wrap(ErrInvalidPhaseParams, "duplicate asset denom")
		}

		if err := sdk.ValidateDenom(asset.Denom); err != nil {
			return errorsmod.Wrapf(ErrInvalidPhaseParams, "invalid asset denom: %v", err)
		}

		if asset.RewardRatio.IsNil() || asset.RewardRatio.IsZero() || asset.RewardRatio.GTE(sdkmath.LegacyOneDec()) {
			return errorsmod.Wrap(ErrInvalidPhaseParams, "invalid asset reward ratio")
		}

		allowedAssets[asset.Denom] = true
	}

	return nil
}
