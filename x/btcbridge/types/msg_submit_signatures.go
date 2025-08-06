package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitSignatures{}

func NewMsgSubmitSignatures(
	sender string,
	txid string,
	signatures []string,
) *MsgSubmitSignatures {
	return &MsgSubmitSignatures{
		Sender:     sender,
		Txid:       txid,
		Signatures: signatures,
	}
}

func (msg *MsgSubmitSignatures) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errorsmod.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Txid) == 0 {
		return errorsmod.Wrap(ErrInvalidTxHash, "tx id cannot be empty")
	}

	if _, err := chainhash.NewHashFromStr(msg.Txid); err != nil {
		return ErrInvalidTxHash
	}

	if len(msg.Signatures) == 0 {
		return errorsmod.Wrap(ErrInvalidSignatures, "signatures cannot be empty")
	}

	for _, signature := range msg.Signatures {
		sigBytes, err := hex.DecodeString(signature)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidSignature, "failed to decode signature")
		}

		if _, err := schnorr.ParseSignature(sigBytes); err != nil {
			return errorsmod.Wrap(ErrInvalidSignature, "invalid schnorr signature")
		}
	}

	return nil
}
