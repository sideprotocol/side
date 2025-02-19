package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitPrice{}

func NewMsgSubmitPrice(sender string, price string) *MsgSubmitPrice {
	return &MsgSubmitPrice{
		Sender: sender,
		Price:  price,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgSubmitPrice) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	_, ok := sdkmath.NewIntFromString(m.Price)
	if !ok {
		return fmt.Errorf("invalid price")
	}

	return nil
}
