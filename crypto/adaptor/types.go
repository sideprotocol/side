package adaptor

import (
	"github.com/btcsuite/btcd/btcec/v2"
)

// scalarSize is the size of an encoded big endian scalar
const scalarSize = 32

// Signature represents the signature
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

// Serialize serializes the signature
func (s *Signature) Serialize() []byte {
	sig := make([]byte, 64)

	rBytes := *s.r.Bytes()
	sBytes := s.s.Bytes()

	copy(sig[0:32], rBytes[:])
	copy(sig[32:64], sBytes[:])

	return sig
}

// SerializeScalar serializes the given scalar
func SerializeScalar(scalar *btcec.ModNScalar) []byte {
	bz := scalar.Bytes()
	return bz[:]
}

// SecretToPubKey gets the serialized public key of the given secret on the secp256k1 curve
func SecretToPubKey(secretBytes []byte) []byte {
	var secret btcec.ModNScalar
	secret.SetByteSlice(secretBytes)

	var result btcec.JacobianPoint
	btcec.ScalarBaseMultNonConst(&secret, &result)

	return btcec.JacobianToByteSlice(result)
}

// NegatePoint negates the given point
func NegatePoint(point *btcec.JacobianPoint) *btcec.JacobianPoint {
	result := *point
	result.Y.Negate(1).Normalize()

	return &result
}
