package bech32

import (
	"fmt"
	"strings"

	// "github.com/cosmos/btcutil/bech32"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/sideprotocol/side/bitcoin"
)

// ConvertAndEncode converts from a base256 encoded byte string to base32 encoded byte string and then to bech32.
func ConvertAndEncode(hrp string, data []byte) (string, error) {

	// use length of hrp to determine if it is an account address
	// check if address is a taproot/sigwit address
	if len(hrp) < 6 {
		if len(data) == 32 { // taproot
			return encodeSegWitAddress(bitcoin.Network.Bech32HRPSegwit, 1, data)
		}
		// segwit address
		bitcoinBech32, err := bech32.Encode(bitcoin.Network.Bech32HRPSegwit, data)
		if IsBitCoinAddr(bitcoinBech32) == "segwit" && err == nil {
			return bitcoinBech32, err
		}
	}

	// other cosmos addresses
	converted, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("encoding bech32 failed: %w", err)
	}
	return bech32.Encode(hrp, converted)
}

// DecodeAndConvert decodes a bech32 encoded string and converts to base256 encoded bytes.
func DecodeAndConvert(bech string) (string, []byte, error) {

	addrType := IsBitCoinAddr(bech)

	if addrType == "taproot" {
		addr, err := btcutil.DecodeAddress(bech, bitcoin.Network)
		if err != nil {
			return "", nil, fmt.Errorf("decoding taproot bech32 failed: %w", err)
		}
		return bitcoin.Network.Bech32HRPSegwit, addr.ScriptAddress(), nil
	} else if addrType == "segwit" {
		hrp, data, err := bech32.Decode(bech)
		if err != nil {
			return "", nil, fmt.Errorf("decoding segwit bech32 failed: %w", err)
		}
		return hrp, data, nil
	}

	hrp, data, err := bech32.Decode(bech)
	if err != nil {
		return "", nil, fmt.Errorf("decoding bech32 failed: %w", err)
	}

	converted, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", nil, fmt.Errorf("decoding bech32 failed: %w", err)
	}
	return hrp, converted, nil

}

func IsBitCoinAddr(bech string) string {
	if strings.HasPrefix(bech, bitcoin.Network.Bech32HRPSegwit+"1q") && len(bech) == 42 {
		return "segwit"
	} else if strings.HasPrefix(bech, bitcoin.Network.Bech32HRPSegwit+"1p") && len(bech) == 62 {
		return "taproot"
	}
	return "cosmos"
}

func encodeSegWitAddress(hrp string, witnessVersion byte, witnessProgram []byte) (string, error) {
	// Group the address bytes into 5 bit groups, as this is what is used to
	// encode each character in the address string.
	converted, err := bech32.ConvertBits(witnessProgram, 8, 5, true)
	if err != nil {
		return "", err
	}

	// Concatenate the witness version and program, and encode the resulting
	// bytes using bech32 encoding.
	combined := make([]byte, len(converted)+1)
	combined[0] = witnessVersion
	copy(combined[1:], converted)

	var bech string
	switch witnessVersion {
	case 0:
		bech, err = bech32.Encode(hrp, combined)

	case 1:
		bech, err = bech32.EncodeM(hrp, combined)

	default:
		return "", fmt.Errorf("unsupported witness version %d",
			witnessVersion)
	}
	if err != nil {
		return "", err
	}

	// Check validity by decoding the created address.
	// version, program, err := decodeSegWitAddress(bech)
	// if err != nil {
	// 	return "", fmt.Errorf("invalid segwit address: %v", err)
	// }

	// if version != witnessVersion || !bytes.Equal(program, witnessProgram) {
	// 	return "", fmt.Errorf("invalid segwit address")
	// }

	return bech, nil
}
