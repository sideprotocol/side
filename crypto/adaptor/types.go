package adaptor

import (
	"github.com/btcsuite/btcd/btcec/v2"
)

// scalarSize is the size of an encoded big endian scalar
const scalarSize = 32

// Signature is same with schnorr.Signature
type Signature struct {
	r btcec.FieldVal
	s btcec.ModNScalar
}

// NewSignature creates a new Signature from bytes
// Assume that the given signature is valid
func NewSignature(sigBytes []byte) *Signature {
	var r btcec.FieldVal
	r.SetByteSlice(sigBytes[0:32])

	var s btcec.ModNScalar
	s.SetByteSlice(sigBytes[32:])

	return &Signature{
		r,
		s,
	}
}

// NegatePoint negates the given point
func NegatePoint(point *btcec.JacobianPoint) *btcec.JacobianPoint {
	result := *point
	result.Y.Negate(1).Normalize()

	return &result
}

// SerializeScalar serializes the given scalar
func SerializeScalar(scalar *btcec.ModNScalar) []byte {
	bz := scalar.Bytes()
	return bz[:]
}
