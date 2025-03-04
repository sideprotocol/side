package adaptor

import (
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Extract extracts the secret from the given adaptor signature and adapted signature
func Extract(adaptorSigBytes []byte, adaptedSigBytes []byte) []byte {
	adaptorR, err := schnorr.ParsePubKey(adaptorSigBytes[0:32])
	if err != nil {
		return nil
	}

	adaptedR, err := schnorr.ParsePubKey(adaptedSigBytes[0:32])
	if err != nil {
		return nil
	}

	var adaptorRPoint, adaptedRPoint secp256k1.JacobianPoint
	adaptorR.AsJacobian(&adaptorRPoint)
	adaptedR.AsJacobian(&adaptedRPoint)

	adaptorSig := NewSignature(adaptorSigBytes)
	adaptedSig := NewSignature(adaptedSigBytes)

	t := adaptedSig.s.Add(adaptorSig.s.Negate())

	switch {
	case verifySecret(t, false, adaptorRPoint, adaptedRPoint):
		return SerializeScalar(t)

	case verifySecret(t.Negate(), true, adaptorRPoint, adaptedRPoint):
		return SerializeScalar(t)

	default:
		return nil
	}
}

// verifySecret returns true if the computed R is correct according to the given secret and parity, false otherwise
func verifySecret(t *secp256k1.ModNScalar, expectOdd bool, adaptorRPoint secp256k1.JacobianPoint, adaptedRPoint secp256k1.JacobianPoint) bool {
	var T secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(t, &T)

	var computedAdaptedRPoint secp256k1.JacobianPoint
	secp256k1.AddNonConst(&adaptorRPoint, &T, &computedAdaptedRPoint)

	adaptedRPoint.ToAffine()
	computedAdaptedRPoint.ToAffine()

	return computedAdaptedRPoint.Y.IsOdd() == expectOdd && computedAdaptedRPoint.X.Equals(&adaptedRPoint.X)
}
