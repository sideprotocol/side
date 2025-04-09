package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgApply{}

func NewMsgApply(borrower string, borrowerPubkey string, maturity int64, poolId string, borrowAmount sdk.Coin, dcmId uint64) *MsgApply {
	return &MsgApply{
		Borrower:       borrower,
		BorrowerPubkey: borrowerPubkey,
		Maturity:       maturity,
		PoolId:         poolId,
		BorrowAmount:   borrowAmount,
		DCMId:          dcmId,
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

	if m.Maturity <= 0 {
		return errorsmod.Wrap(ErrInvalidMaturity, "maturity must be greater than 0")
	}

	if len(m.PoolId) == 0 {
		return errorsmod.Wrap(ErrInvalidPoolId, "empty pool id")
	}

	if !m.BorrowAmount.IsValid() || !m.BorrowAmount.IsPositive() {
		return errorsmod.Wrap(ErrInvalidAmount, "borrowed amount must be positive")
	}

	return nil
}
