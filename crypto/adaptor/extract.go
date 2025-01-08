package adaptor

import (
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Extract extracts the secret from the given adapted signature and adaptor signature
func Extract(adaptedSigBytes []byte, adaptorSigBytes []byte) []byte {
	adaptedR, err := schnorr.ParsePubKey(adaptedSigBytes[0:32])
	if err != nil {
		return nil
	}

	adaptorR, err := schnorr.ParsePubKey(adaptorSigBytes[0:32])
	if err != nil {
		return nil
	}

	var adaptedRPoint, adaptorRPoint secp256k1.JacobianPoint
	adaptedR.AsJacobian(&adaptedRPoint)
	adaptorR.AsJacobian(&adaptorRPoint)

	var rPointSub, rPointAdd secp256k1.JacobianPoint
	secp256k1.AddNonConst(&adaptedRPoint, NegatePoint(&adaptorRPoint), &rPointSub)
	secp256k1.AddNonConst(&adaptedRPoint, &adaptorRPoint, &rPointAdd)

	adaptedSig := NewSignature(adaptedSigBytes)
	adaptorSig := NewSignature(adaptorSigBytes)

	t := adaptedSig.s.Add(adaptorSig.s.Negate())

	var T secp256k1.JacobianPoint
	secp256k1.ScalarBaseMultNonConst(t, &T)

	switch T {
	case rPointSub:
		return SerializeScalar(t)
	case rPointAdd:
		return SerializeScalar(t.Negate())
	default:
		return nil
	}
}
