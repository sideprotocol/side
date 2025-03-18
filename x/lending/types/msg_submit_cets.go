package types

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitCets{}

func NewMsgSubmitCets(borrower string, loanId string, depositTx string, liquidationCet string, liquidationAdaptorSignatures []string, defaultLiquidationAdaptorSignatures []string, repaymentCet string, repaymentSignatures []string) *MsgSubmitCets {
	return &MsgSubmitCets{
		Borrower:                            borrower,
		LoanId:                              loanId,
		DepositTx:                           depositTx,
		LiquidationCet:                      liquidationCet,
		LiquidationAdaptorSignatures:        liquidationAdaptorSignatures,
		DefaultLiquidationAdaptorSignatures: defaultLiquidationAdaptorSignatures,
		RepaymentCet:                        repaymentCet,
		RepaymentSignatures:                 repaymentSignatures,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgSubmitCets) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Borrower); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	if _, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(m.DepositTx)), true); err != nil {
		return ErrInvalidDepositTx
	}

	liquidationCet, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(m.LiquidationCet)), true)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidCET, "failed to deserialize liquidation cet: %v", err)
	}

	if len(m.LiquidationAdaptorSignatures) != len(liquidationCet.Inputs) {
		return errorsmod.Wrap(ErrInvalidAdaptorSignatures, "incorrect liquidation adaptor signature number")
	}

	if len(m.DefaultLiquidationAdaptorSignatures) != len(liquidationCet.Inputs) {
		return errorsmod.Wrap(ErrInvalidAdaptorSignatures, "incorrect default liquidation adaptor signature number")
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

	for _, signature := range m.DefaultLiquidationAdaptorSignatures {
		adaptorSigBytes, err := hex.DecodeString(signature)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidAdaptorSignature, "failed to decode adaptor signature")
		}

		if _, err := schnorr.ParseSignature(adaptorSigBytes); err != nil {
			return ErrInvalidAdaptorSignature
		}
	}

	repaymentCet, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(m.RepaymentCet)), true)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidCET, "failed to deserialize repayment cet: %v", err)
	}

	if len(m.RepaymentSignatures) != len(repaymentCet.Inputs) {
		return errorsmod.Wrap(ErrInvalidSignatures, "incorrect repayment signature number")
	}

	for _, signature := range m.RepaymentSignatures {
		sigBytes, err := hex.DecodeString(signature)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidSignature, "failed to decode signature")
		}

		if _, err := schnorr.ParseSignature(sigBytes); err != nil {
			return ErrInvalidSignature
		}
	}

	return nil
}
