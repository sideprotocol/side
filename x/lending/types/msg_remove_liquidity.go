package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRemoveLiquidity{}

func NewMsgRemoveLiquidity(lender string, shares sdk.Coin) *MsgRemoveLiquidity {
	return &MsgRemoveLiquidity{
		Lender: lender,
		Shares: shares,
	}
}

// ValidateBasic performs basic MsgRemoveLiquidity message validation.
func (m *MsgRemoveLiquidity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Lender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if !m.Shares.IsValid() || !m.Shares.IsPositive() {
		return errorsmod.Wrap(ErrInvalidAmount, "shares must be positive")
	}

	return nil
}
