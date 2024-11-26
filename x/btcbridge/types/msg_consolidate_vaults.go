package types

import (
	"lukechampine.com/uint128"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgConsolidateVaults{}

// ValidateBasic performs basic MsgConsolidateVaults message validation.
func (m *MsgConsolidateVaults) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if m.BtcConsolidation == nil && len(m.RunesConsolidations) == 0 {
		return errorsmod.Wrap(ErrInvalidConsolidation, "neither btc nor runes consolidation provided")
	}

	if m.BtcConsolidation != nil {
		if err := ensureBtcConsolidation(m.BtcConsolidation); err != nil {
			return err
		}
	}

	if len(m.RunesConsolidations) != 0 {
		if err := ensureRunesConsolidations(m.RunesConsolidations); err != nil {
			return err
		}
	}

	return nil
}

// ensureBtcConsolidation checks the given btc consolidation
func ensureBtcConsolidation(consolidation *BtcConsolidation) error {
	if consolidation.TargetThreshold <= 0 {
		return errorsmod.Wrap(ErrInvalidConsolidation, "btc target threshold must be greater than 0")
	}

	return nil
}

// ensureRunesConsolidations checks the given runes consolidations
func ensureRunesConsolidations(consolidations []*RunesConsolidation) error {
	for _, c := range consolidations {
		var id RuneId
		err := id.FromString(c.RuneId)
		if err != nil {
			return err
		}

		threshold, err := uint128.FromString(c.TargetThreshold)
		if err != nil || threshold.IsZero() {
			return errorsmod.Wrap(ErrInvalidConsolidation, "invalid runes target threshold")
		}
	}

	return nil
}
