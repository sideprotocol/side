package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRemoveLiquidity{}

func NewMsgRemoveLiquidity(lender string, amount sdk.Coin) *MsgRemoveLiquidity {
	return &MsgRemoveLiquidity{
		Lender: lender,
		Shares: &amount,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgRemoveLiquidity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Lender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if m.Shares.Amount.LTE(math.NewInt(0)) {
		return ErrInvalidLiquidity
	}

	return nil
}
