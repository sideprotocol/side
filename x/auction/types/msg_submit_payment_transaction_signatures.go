package types

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

var _ sdk.Msg = &MsgSubmitPaymentTransactionSignatures{}

func NewMsgSubmitPaymentTransactionSignatures(
	relayer string,
	auctionId uint64,
	signatures []string,
) *MsgSubmitPaymentTransactionSignatures {
	return &MsgSubmitPaymentTransactionSignatures{
		Relayer:    relayer,
		AuctionId:  auctionId,
		Signatures: signatures,
	}
}

// ValidateBasic performs basic MsgSubmitPaymentTransactionSignatures message validation.
func (m *MsgSubmitPaymentTransactionSignatures) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Relayer); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.Signatures) == 0 {
		return errorsmod.Wrap(ErrInvalidSignatures, "signatures can not be empty")
	}

	for _, sig := range m.Signatures {
		sigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return errorsmod.Wrapf(ErrInvalidSignature, "%v", err)
		}

		if _, err := schnorr.ParseSignature(sigBytes); err != nil {
			return errorsmod.Wrapf(ErrInvalidSignature, "%v", err)
		}
	}

	return nil
}
