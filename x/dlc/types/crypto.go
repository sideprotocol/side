package types

import (
	"crypto/sha256"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

// Sha256 returns the SHA256 hash of the given data
func Sha256(data []byte) []byte {
	hash := sha256.Sum256(data)

	return hash[:]
}

// VerifySchnorrSignature verifies the provided schnorr signature against the given hash and public key
func VerifySchnorrSignature(sig []byte, sigHash []byte, pubKeyBytes []byte) bool {
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
