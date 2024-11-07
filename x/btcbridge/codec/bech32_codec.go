package codec

import (
	"errors"
	"strings"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type bech32Codec struct {
	nativeBech32Prefix  string
	bitcoinBech32Prefix string
}

var _ address.Codec = &bech32Codec{}

// NewBech32Codec creates a new address codec
func NewBech32Codec(nativeBech32Prefix string, bitcoinBech32Prefix string) address.Codec {
	return bech32Codec{nativeBech32Prefix, bitcoinBech32Prefix}
}

// StringToBytes encodes text to bytes
func (bc bech32Codec) StringToBytes(text string) ([]byte, error) {
	if len(strings.TrimSpace(text)) == 0 {
		return []byte{}, errors.New("empty address string is not allowed")
	}

	hrp, bz, err := bech32.DecodeAndConvert(text)
	if err != nil {
		return nil, err
	}

	if hrp != bc.nativeBech32Prefix && hrp != bc.bitcoinBech32Prefix {
		return nil, errorsmod.Wrapf(sdkerrors.ErrLogic, "hrp does not match bech32 prefixes: expected '%s or %s' got '%s'", bc.nativeBech32Prefix, bc.bitcoinBech32Prefix, hrp)
	}

	if err := sdk.VerifyAddressFormat(bz); err != nil {
		return nil, err
	}

	return bz, nil
}

// BytesToString decodes bytes to text
func (bc bech32Codec) BytesToString(bz []byte) (string, error) {
	bech32Prefix := bc.nativeBech32Prefix
	if isBitcoinAddress(bz) {
		bech32Prefix = bc.bitcoinBech32Prefix
	}

	text, err := bech32.ConvertAndEncode(bech32Prefix, bz)
	if err != nil {
		return "", err
	}

	return text, nil
}

// isBitcoinAddress returns true if the given address is segwit or taproot, false otherwise
func isBitcoinAddress(address []byte) bool {
	return isSegwitAddress(address) || isTaprootAddress(address)
}

// isSegwitAddress returns true if the given address is segwit, false otherwise
func isSegwitAddress(address []byte) bool {
	return len(address) == 33 // bech32 decoded address length
}

// isTaprootAddress returns true if the given address is taproot, false otherwise
func isTaprootAddress(address []byte) bool {
	return len(address) == 32 // only with taproot output key
}
