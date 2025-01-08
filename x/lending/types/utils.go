package types

import "github.com/btcsuite/btcd/btcec/v2"

// SecretToPubKey gets the serialized public key of the given secret on the secp256k1 curve
// Assume that the secret is 256-bit bytes
func SecretToPubKey(secretBytes []byte) []byte {
	var secret btcec.ModNScalar
	_ = secret.SetByteSlice(secretBytes)

	var result btcec.JacobianPoint
	btcec.ScalarBaseMultNonConst(&secret, &result)

	return btcec.JacobianToByteSlice(result)
}
