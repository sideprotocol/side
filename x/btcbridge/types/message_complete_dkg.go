package types

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCompleteDKG = "complete_dkg"

func NewMsgCompleteDKG(
	sender string,
	id uint64,
	vaults []string,
	validator string,
	signature string,
) *MsgCompleteDKG {
	return &MsgCompleteDKG{
		Sender:    sender,
		Id:        id,
		Vaults:    vaults,
		Validator: validator,
		Signature: signature,
	}
}

// Route returns the route of MsgCompleteDKG.
func (msg *MsgCompleteDKG) Route() string {
	return RouterKey
}

// Type returns the type of MsgCompleteDKG.
func (msg *MsgCompleteDKG) Type() string {
	return TypeMsgCompleteDKG
}

// GetSignBytes implements the LegacyMsg interface.
func (m MsgCompleteDKG) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCompleteDKG message.
func (m *MsgCompleteDKG) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic MsgCompleteDKG message validation.
func (m *MsgCompleteDKG) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(err, "invalid sender address")
	}

	if len(m.Vaults) == 0 {
		return ErrInvalidDKGCompletionRequest
	}

	if _, err := sdk.ConsAddressFromHex(m.Validator); err != nil {
		return ErrInvalidDKGCompletionRequest
	}

	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil {
		return ErrInvalidDKGCompletionRequest
	}

	if len(sigBytes) != ed25519.SignatureSize {
		return ErrInvalidDKGCompletionRequest
	}

	return nil
}
