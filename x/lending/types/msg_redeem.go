package types

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRedeem{}

func NewMsgRedeem(borrower string, loanId string, tx string, signatures []string) *MsgRedeem {
	return &MsgRedeem{
		Borrower:   borrower,
		LoanId:     loanId,
		Tx:         tx,
		Signatures: signatures,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgRedeem) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Borrower); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(m.Tx)), true)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidTx, "failed to deserialize tx: %v", err)
	}

	if len(m.Signatures) != len(p.Inputs) {
		return errorsmod.Wrap(ErrInvalidSignatures, "incorrect signature number")
	}

	for _, signature := range m.Signatures {
		sigBytes, err := hex.DecodeString(signature)
		if err != nil {
			return ErrInvalidSignature
		}

		if _, err := schnorr.ParseSignature(sigBytes); err != nil {
			return ErrInvalidSignature
		}
	}

	return nil
}
