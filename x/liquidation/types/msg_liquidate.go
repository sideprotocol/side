package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgLiquidate{}

func NewMsgLiquidate(
	liquidator string,
	liquidationId uint64,
	debtAmount sdk.Coin,
) *MsgLiquidate {
	return &MsgLiquidate{
		Liquidator:    liquidator,
		LiquidationId: liquidationId,
		DebtAmount:    debtAmount,
	}
}

// ValidateBasic performs basic MsgLiquidate message validation.
func (m *MsgLiquidate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Liquidator); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if !IsValidBtcAddress(m.Liquidator) {
		return errorsmod.Wrap(ErrInvalidSender, "liquidator address must be a valid btc address")
	}

	if !m.DebtAmount.IsValid() || !m.DebtAmount.IsPositive() {
		return errorsmod.Wrap(ErrInvalidAmount, "invalid debt amount")
	}

	return nil
}
