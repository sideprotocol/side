package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgApply{}

func NewMsgApply(borrower string, borrowerPubkey string, borrowerAuthPubkey string, poolId string, borrowAmount sdk.Coin, maturity int64, dcmId uint64, referralCode string) *MsgApply {
	return &MsgApply{
		Borrower:           borrower,
		BorrowerPubkey:     borrowerPubkey,
		BorrowerAuthPubkey: borrowerAuthPubkey,
		PoolId:             poolId,
		BorrowAmount:       borrowAmount,
		Maturity:           maturity,
		DCMId:              dcmId,
		ReferralCode:       referralCode,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgApply) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Borrower); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	pubKeyBytes, err := hex.DecodeString(m.BorrowerPubkey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode borrower public key")
	}

	if _, err := schnorr.ParsePubKey(pubKeyBytes); err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "invalid borrower public key")
	}

	authPubKeyBytes, err := hex.DecodeString(m.BorrowerAuthPubkey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode borrower auth public key")
	}

	if _, err := schnorr.ParsePubKey(authPubKeyBytes); err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "invalid borrower auth public key")
	}

	if len(m.PoolId) == 0 {
		return errorsmod.Wrap(ErrInvalidPoolId, "empty pool id")
	}

	if !m.BorrowAmount.IsValid() || !m.BorrowAmount.IsPositive() {
		return errorsmod.Wrap(ErrInvalidAmount, "borrow amount must be positive")
	}

	if m.Maturity <= 0 {
		return errorsmod.Wrap(ErrInvalidMaturity, "maturity must be greater than 0")
	}

	if len(m.ReferralCode) != 0 && !ReferralCodeRegex.MatchString(m.ReferralCode) {
		return errorsmod.Wrap(ErrInvalidReferralCode, "referral code must be 8 alphanumeric characters")
	}

	return nil
}
