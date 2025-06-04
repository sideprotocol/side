package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRemoveLiquidity{}

func NewMsgRemoveLiquidity(lender string, sTokens sdk.Coin) *MsgRemoveLiquidity {
	return &MsgRemoveLiquidity{
		Lender:  lender,
		STokens: sTokens,
	}
}

// ValidateBasic performs basic MsgRemoveLiquidity message validation.
func (m *MsgRemoveLiquidity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Lender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if !strings.HasPrefix(m.STokens.Denom, S_TOKEN_DENOM_PREFIX) {
		return errorsmod.Wrap(ErrInvalidAmount, "invalid sToken denom")
	}

	if !m.STokens.IsValid() || !m.STokens.IsPositive() {
		return errorsmod.Wrap(ErrInvalidAmount, "sTokens must be positive")
	}

	return nil
}
