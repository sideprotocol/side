package schnorr

import (
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

// Verify verifies the provided schnorr signature against the given msg and public key
func Verify(sigBytes []byte, msg []byte, pubKeyBytes []byte) bool {
	signature, err := schnorr.ParseSignature(sigBytes)
	if err != nil {
		return false
	}

	pubKey, err := schnorr.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false
	}

	return signature.Verify(msg, pubKey)
}
