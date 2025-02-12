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

func NewMsgApply(borrower string, borrowerPubkey string, hashLoanSecret string, maturityTime int64, finalTimeout int64) *MsgApply {
	return &MsgApply{
		Borrower:       borrower,
		BorrowerPubkey: borrowerPubkey,
		LoanSecretHash: hashLoanSecret,
		MaturityTime:   maturityTime,
		FinalTimeout:   finalTimeout,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgApply) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Borrower); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	pubKeyBytes, err := hex.DecodeString(m.BorrowerPubkey)
	if err != nil {
		return ErrInvalidBorrowerPubkey
	}

	if _, err := schnorr.ParsePubKey(pubKeyBytes); err != nil {
		return ErrInvalidBorrowerPubkey
	}

	if m.MaturityTime <= 0 {
		return ErrInvalidMaturityTime
	}

	if m.MaturityTime <= m.FinalTimeout {
		return ErrInvalidFinalTimeout
	}

	if secretHashBytes, err := hex.DecodeString(m.LoanSecretHash); err != nil || len(secretHashBytes) != LoanSecretHashLength {
		return ErrInvalidLoanSecretHash
	}

	_, err = psbt.NewFromRawBytes(bytes.NewReader([]byte(m.DepositTx)), true)
	if err != nil {
		return ErrInvalidDepositTx
	}

	return nil
}
