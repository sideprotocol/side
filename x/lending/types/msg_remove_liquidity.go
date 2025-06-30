package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRemoveLiquidity{}

func NewMsgRemoveLiquidity(lender string, yTokens sdk.Coin) *MsgRemoveLiquidity {
	return &MsgRemoveLiquidity{
		Lender:  lender,
		YTokens: yTokens,
	}
}

// ValidateBasic performs basic MsgRemoveLiquidity message validation.
func (m *MsgRemoveLiquidity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Lender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if !strings.HasPrefix(m.YTokens.Denom, Y_TOKEN_DENOM_PREFIX) {
		return errorsmod.Wrap(ErrInvalidAmount, "invalid yToken denom")
	}

	if !m.YTokens.IsValid() || !m.YTokens.IsPositive() {
		return errorsmod.Wrap(ErrInvalidAmount, "yTokens must be positive")
	}

	return nil
}
