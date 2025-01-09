package adaptor

import (
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Adapt adapts the given adaptor signature with the specified secret
// Asume that the given adaptor signature is valid
func Adapt(sigBytes []byte, secretBytes []byte) []byte {
	rPoint, _ := schnorr.ParsePubKey(sigBytes[0:32])
	var R secp256k1.JacobianPoint
	rPoint.AsJacobian(&R)

	var s secp256k1.ModNScalar
	s.SetByteSlice(sigBytes[32:64])

	var secret secp256k1.ModNScalar
	secret.SetByteSlice(secretBytes)

	var adaptorPoint secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(&secret, &adaptorPoint)

	var adaptedR secp256k1.JacobianPoint
	secp256k1.AddNonConst(&R, &adaptorPoint, &adaptedR)
	adaptedR.ToAffine()

	var adaptedS secp256k1.ModNScalar
	if !adaptedR.Y.IsOdd() {
		adaptedS = *s.Add(&secret)
	} else {
		adaptedS = *s.Add(secret.Negate())
	}

	adaptedSig := Signature{
		r: adaptedR.X,
		s: adaptedS,
	}

	return adaptedSig.Serialize()
}
