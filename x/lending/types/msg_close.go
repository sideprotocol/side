package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgClose{}

func NewMsgClose(relayer string, loanId string, signature string) *MsgClose {
	return &MsgClose{
		Relayer:   relayer,
		LoanId:    loanId,
		Signature: signature,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgClose) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Relayer); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil {
		return ErrInvalidSignature
	}

	if _, err := schnorr.ParseSignature(sigBytes); err != nil {
		return ErrInvalidSignature
	}

	return nil
}
