package types

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCompleteDKG{}

func NewMsgCompleteDKG(
	sender string,
	id uint64,
	vaults []string,
	consAddress string,
	signature string,
) *MsgCompleteDKG {
	return &MsgCompleteDKG{
		Sender:           sender,
		Id:               id,
		Vaults:           vaults,
		ConsensusAddress: consAddress,
		Signature:        signature,
	}
}

// ValidateBasic performs basic MsgCompleteDKG message validation.
func (m *MsgCompleteDKG) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.Vaults) == 0 {
		return errorsmod.Wrap(ErrInvalidDKGCompletionRequest, "vaults can not be empty")
	}

	vaults := make(map[string]bool)
	for _, v := range m.Vaults {
		_, err := sdk.AccAddressFromBech32(v)
		if err != nil || vaults[v] {
			return errorsmod.Wrap(ErrInvalidDKGCompletionRequest, "invalid vault")
		}

		vaults[v] = true
	}

	if _, err := sdk.ConsAddressFromHex(m.ConsensusAddress); err != nil {
		return errorsmod.Wrap(ErrInvalidDKGCompletionRequest, "invalid consensus address")
	}

	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidDKGCompletionRequest, "failed to decode signature")
	}

	if len(sigBytes) != ed25519.SignatureSize {
		return errorsmod.Wrap(ErrInvalidDKGCompletionRequest, "invalid signature size")
	}

	return nil
}
