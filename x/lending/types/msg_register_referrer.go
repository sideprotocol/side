package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRegisterReferrer{}

// ValidateBasic performs basic MsgRegisterReferrer message validation.
func (m *MsgRegisterReferrer) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if len(m.Name) > MaxReferrerNameLength {
		return errorsmod.Wrapf(ErrInvalidReferrer, "referrer name length %d exceeds the allowed maximum length %d", len(m.Name), MaxReferrerNameLength)
	}

	if !ReferralCodeRegex.MatchString(m.ReferralCode) {
		return errorsmod.Wrap(ErrInvalidReferralCode, "referral code must be 8 alphanumeric characters")
	}

	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		return errorsmod.Wrap(err, "invalid referrer address")
	}

	if m.ReferralFeeFactor.IsNegative() || m.ReferralFeeFactor.GT(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidReferralFeeFactor, "referral fee factor must be between [0, 1]")
	}

	return nil
}
