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
// Assume that the given byte slice is valid
func NewSignature(sigBytes []byte) *Signature {
	var r btcec.FieldVal
	_ = r.SetByteSlice(sigBytes[0:32])

	var s btcec.ModNScalar
	_ = s.SetByteSlice(sigBytes[32:])

	return &Signature{
		r,
		s,
	}
}
