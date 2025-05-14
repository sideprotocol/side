package adaptor

import (
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Adapt adapts the given adaptor signature with the specified secret
// Asume that the given adaptor signature is valid
func Adapt(sigBytes []byte, secretBytes []byte) []byte {
	sig, _ := ParseSignature(sigBytes)

	var secret secp256k1.ModNScalar
	secret.SetByteSlice(secretBytes)

	var adaptedS secp256k1.ModNScalar

	if sig.IsROdd() {
		adaptedS = *sig.s.Add(secret.Negate())
	} else {
		adaptedS = *sig.s.Add(&secret)
	}

	adaptedSig := schnorr.NewSignature(&sig.r, &adaptedS)

	return adaptedSig.Serialize()
}
