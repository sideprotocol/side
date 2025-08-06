package types

import (
	"bytes"
	"encoding/hex"
	"slices"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/bitcoin/crypto/adaptor"
)

var _ sdk.Msg = &MsgSubmitCets{}

func NewMsgSubmitCets(borrower string, loanId string, depositTxs []string, liquidationCet string, liquidationAdaptorSignatures []string, defaultLiquidationAdaptorSignatures []string, repaymentCet string, repaymentSignatures []string) *MsgSubmitCets {
	return &MsgSubmitCets{
		Borrower:                            borrower,
		LoanId:                              loanId,
		DepositTxs:                          depositTxs,
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

	if len(m.DepositTxs) == 0 {
		return errorsmod.Wrap(ErrInvalidDepositTxs, "deposit txs cannot be empty")
	}

	depositTxHashes := []string{}

	for _, depositTx := range m.DepositTxs {
		if p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(depositTx)), true); err != nil {
			return ErrInvalidDepositTx
		} else {
			depositTxHashes = append(depositTxHashes, p.UnsignedTx.TxHash().String())
		}
	}

	liquidationCet, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(m.LiquidationCet)), true)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidCET, "failed to deserialize liquidation cet: %v", err)
	}

	for _, txIn := range liquidationCet.UnsignedTx.TxIn {
		if !slices.Contains(depositTxHashes, txIn.PreviousOutPoint.Hash.String()) {
			return errorsmod.Wrapf(ErrInvalidCET, "invalid previous tx hash in liquidation cet")
		}
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

		if _, err := adaptor.ParseSignature(adaptorSigBytes); err != nil {
			return ErrInvalidAdaptorSignature
		}
	}

	for _, signature := range m.DefaultLiquidationAdaptorSignatures {
		adaptorSigBytes, err := hex.DecodeString(signature)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidAdaptorSignature, "failed to decode adaptor signature")
		}

		if _, err := adaptor.ParseSignature(adaptorSigBytes); err != nil {
			return ErrInvalidAdaptorSignature
		}
	}

	repaymentCet, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(m.RepaymentCet)), true)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidCET, "failed to deserialize repayment cet: %v", err)
	}

	for _, txIn := range repaymentCet.UnsignedTx.TxIn {
		if !slices.Contains(depositTxHashes, txIn.PreviousOutPoint.Hash.String()) {
			return errorsmod.Wrapf(ErrInvalidCET, "invalid previous tx hash in repayment cet")
		}
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
