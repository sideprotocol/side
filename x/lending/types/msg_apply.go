package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgApply{}

func NewMsgApply(borrower string, borrowerPubkey string, hashLoanSecret string, maturityTime int64, finalTimeout int64, poolId string, borrowAmount sdk.Coin, agencyId uint64) *MsgApply {
	return &MsgApply{
		Borrower:       borrower,
		BorrowerPubkey: borrowerPubkey,
		LoanSecretHash: hashLoanSecret,
		MaturityTime:   maturityTime,
		FinalTimeout:   finalTimeout,
		PoolId:         poolId,
		BorrowAmount:   &borrowAmount,
		AgencyId:       agencyId,
	}
}

// ValidateBasic performs basic message validation.
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

	if m.FinalTimeout <= m.MaturityTime {
		return ErrInvalidFinalTimeout
	}

	if secretHashBytes, err := hex.DecodeString(m.LoanSecretHash); err != nil || len(secretHashBytes) != LoanSecretHashLength {
		return ErrInvalidLoanSecretHash
	}

	if !m.BorrowAmount.IsValid() || !m.BorrowAmount.IsPositive() {
		return ErrInvalidAmount
	}

	return nil
}
