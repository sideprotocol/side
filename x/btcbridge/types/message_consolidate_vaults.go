package types

import (
	"lukechampine.com/uint128"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgConsolidateVaults = "consolidate_vaults"

// Route returns the route of MsgConsolidateVaults.
func (msg *MsgConsolidateVaults) Route() string {
	return RouterKey
}

// Type returns the type of MsgConsolidateVaults.
func (msg *MsgConsolidateVaults) Type() string {
	return TypeMsgConsolidateVaults
}

// GetSignBytes implements the LegacyMsg interface.
func (m MsgConsolidateVaults) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgConsolidateVaults message.
func (m *MsgConsolidateVaults) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic MsgConsolidateVaults message validation.
func (m *MsgConsolidateVaults) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if m.Switch && (m.BtcConsolidation != nil || len(m.RunesConsolidations) != 0) {
		return ErrInvalidConsolidation
	}

	if !m.Switch && m.BtcConsolidation == nil && len(m.RunesConsolidations) == 0 {
		return ErrInvalidConsolidation
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
		return ErrInvalidBtcConsolidation
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

		threshold, _ := uint128.FromString(c.TargetThreshold)
		if threshold.IsZero() {
			return ErrInvalidRunesConsolidation
		}
	}

	return nil
}
