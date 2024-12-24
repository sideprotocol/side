package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgBid{}

func NewMsgBid(
	sender string,
	auctionId uint64,
	price int64,
	amount sdk.Coin,
) *MsgBid {
	return &MsgBid{
		Sender:    sender,
		AuctionId: auctionId,
		Price:     price,
		Amount:    amount,
	}
}

// ValidateBasic performs basic MsgBid message validation.
func (m *MsgBid) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if m.Price <= 0 {
		return errorsmod.Wrap(ErrInvalidBid, "price must be greater than 0")
	}

	if !m.Amount.IsValid() || !m.Amount.IsPositive() {
		return errorsmod.Wrap(ErrInvalidBid, "invalid amount")
	}

	return nil
}
