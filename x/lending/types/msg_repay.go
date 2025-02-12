package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRepay{}

func NewMsgRepay(borrower string, loanId string, adaptorPoint string) *MsgRepay {
	return &MsgRepay{
		Borrower:     borrower,
		AdaptorPoint: adaptorPoint,
		LoanId:       loanId,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgRepay) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Borrower); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	adaptorPointBytes, err := hex.DecodeString(m.AdaptorPoint)
	if err != nil {
		return ErrInvalidAdaptorPoint
	}

	if _, err = btcec.ParsePubKey(adaptorPointBytes); err != nil {
		return ErrInvalidAdaptorPoint
	}

	return nil
}
