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

func NewMsgSubmitLiquidationCet(borrower string, loanId string, eventId uint64, depositTx string, liquidationCet string, liquidationAdaptorSignature string) *MsgSubmitLiquidationCet {
	return &MsgSubmitLiquidationCet{
		Borrower:                    borrower,
		LoanId:                      loanId,
		EventId:                     eventId,
		DepositTx:                   depositTx,
		LiquidationCet:              liquidationCet,
		LiquidationAdaptorSignature: liquidationAdaptorSignature,
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

	_, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(m.DepositTx)), true)
	if err != nil {
		return ErrInvalidDepositTx
	}

	_, err = psbt.NewFromRawBytes(bytes.NewReader([]byte(m.LiquidationCet)), true)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidCET, "failed to deserialize liquidation cet: %v", err)
	}

	adaptorSigBytes, err := hex.DecodeString(m.LiquidationAdaptorSignature)
	if err != nil {
		return ErrInvalidAdaptorSignature
	}

	if _, err := schnorr.ParseSignature(adaptorSigBytes); err != nil {
		return ErrInvalidAdaptorSignature
	}

	return nil
}
