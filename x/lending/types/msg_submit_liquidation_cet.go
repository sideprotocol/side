package types

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgApply{}

func NewMsgSubmitLiquidationCet(borrower string, loanId string, eventId uint64, depositTx string, liquidationCet string, liquidationAdaptorSignatures []string) *MsgSubmitLiquidationCet {
	return &MsgSubmitLiquidationCet{
		Borrower:                     borrower,
		LoanId:                       loanId,
		EventId:                      eventId,
		DepositTx:                    depositTx,
		LiquidationCet:               liquidationCet,
		LiquidationAdaptorSignatures: liquidationAdaptorSignatures,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgSubmitLiquidationCet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Borrower); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	if _, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(m.DepositTx)), true); err != nil {
		return ErrInvalidDepositTx
	}

	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(m.LiquidationCet)), true)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidCET, "failed to deserialize liquidation cet: %v", err)
	}

	if len(m.LiquidationAdaptorSignatures) != len(p.Inputs) {
		return errorsmod.Wrap(ErrInvalidAdaptorSignatures, "incorrect signature number")
	}

	for _, signature := range m.LiquidationAdaptorSignatures {
		adaptorSigBytes, err := hex.DecodeString(signature)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidAdaptorSignature, "failed to decode adaptor signature")
		}

		if _, err := schnorr.ParseSignature(adaptorSigBytes); err != nil {
			return ErrInvalidAdaptorSignature
		}
	}

	return nil
}
