package types

import (
	"crypto/sha256"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

// SHA256 returns the SHA256 hash of the given data
func SHA256(data []byte) []byte {
	hash := sha256.Sum256(data)

	return hash[:]
}

// VerifySignature verifies the provided signature against the given hash and public key
func VerifySignature(sig []byte, sigHash []byte, pubKeyBytes []byte) bool {
	signature, err := schnorr.ParseSignature(sig)
	if err != nil {
		return false
	}

	pubKey, err := schnorr.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false
	}

	return signature.Verify(sigHash, pubKey)
}
