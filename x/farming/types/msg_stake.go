package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgStake{}

func NewMsgStake(staker string, amount sdk.Coin, lockDuration time.Duration) *MsgStake {
	return &MsgStake{
		Staker:       staker,
		Amount:       amount,
		LockDuration: lockDuration,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgStake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Staker); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if !m.Amount.IsValid() || !m.Amount.IsPositive() {
		return errorsmod.Wrap(ErrInvalidAmount, "amount must be positive")
	}

	if m.LockDuration <= 0 {
		return ErrInvalidLockDuration
	}

	return nil
}
