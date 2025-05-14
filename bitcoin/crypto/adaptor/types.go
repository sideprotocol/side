package adaptor

import (
	"errors"

	"github.com/btcsuite/btcd/btcec/v2"
)

const (
	// SignatureSize is the size of the adaptor signature
	SignatureSize = 65

	// scalarSize is the size of an encoded big endian scalar
	scalarSize = 32
)

var (
	ErrInvalidSignatureSize = errors.New("invalid adaptor signature size")
)

// Signature represents the adaptor signature
type Signature struct {
	rParity byte
	r       btcec.FieldVal
	s       btcec.ModNScalar
}

// NewSignature creates a new Signature from bytes
// Assume that the given signature is valid
func NewSignature(sigBytes []byte) *Signature {
	rParity := sigBytes[0]

	var r btcec.FieldVal
	r.SetByteSlice(sigBytes[1:33])

	var s btcec.ModNScalar
	s.SetByteSlice(sigBytes[33:])

	return &Signature{
		rParity,
		r,
		s,
	}
}

// IsROdd returns true if the r point is odd, false otherwise
func (s Signature) IsROdd() bool {
	return s.rParity == byte(3)
}

// Serialize serializes the signature
func (s Signature) Serialize() []byte {
	sig := make([]byte, 65)

	rBytes := s.r.Bytes()
	sBytes := s.s.Bytes()

	sig[0] = s.rParity
	copy(sig[1:33], rBytes[:])
	copy(sig[33:], sBytes[:])

	return sig
}

// ParseSignature parses the signature
func ParseSignature(sigBytes []byte) (*Signature, error) {
	if len(sigBytes) != SignatureSize {
		return nil, ErrInvalidSignatureSize
	}

	_, err := btcec.ParsePubKey(sigBytes[0:33])
	if err != nil {
		return nil, err
	}

	return NewSignature(sigBytes), nil
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
	result := new(btcec.JacobianPoint)

	result.X = point.X
	result.Y = *point.Y.Negate(1).Normalize()
	result.Z = point.Z

	return result
}
