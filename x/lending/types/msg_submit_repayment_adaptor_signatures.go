package types

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/adaptor"
)

var _ sdk.Msg = &MsgSubmitRepaymentAdaptorSignatures{}

func NewMsgSubmitRepaymentAdaptorSignatures(sender string, loanId string, adaptorSignatures []string) *MsgSubmitRepaymentAdaptorSignatures {
	return &MsgSubmitRepaymentAdaptorSignatures{
		Sender:            sender,
		LoanId:            loanId,
		AdaptorSignatures: adaptorSignatures,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgSubmitRepaymentAdaptorSignatures) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	if len(m.AdaptorSignatures) == 0 {
		return errorsmod.Wrap(ErrInvalidAdaptorSignatures, "adaptor signatures can not be empty")
	}

	for _, sig := range m.AdaptorSignatures {
		adaptorSigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidAdaptorSignature, "failed to decode adaptor signature")
		}

		if _, err := adaptor.ParseSignature(adaptorSigBytes); err != nil {
			return ErrInvalidAdaptorSignature
		}
	}

	return nil
}
