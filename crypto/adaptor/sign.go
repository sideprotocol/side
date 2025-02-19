package adaptor

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

var (
	// rfc6979ExtraDataV0 is the extra data to feed to RFC6979 when
	// generating the deterministic nonce for the BIP-340 scheme.  This
	// ensures the same nonce is not generated for the same message and key
	// as for other signing algorithms such as ECDSA.
	//
	// It is equal to SHA-256([]byte("BIP-340")).
	rfc6979ExtraDataV0 = [32]uint8{
		0xa3, 0xeb, 0x4c, 0x18, 0x2f, 0xae, 0x7e, 0xf4,
		0xe8, 0x10, 0xc6, 0xee, 0x13, 0xb0, 0xe9, 0x26,
		0x68, 0x6d, 0x71, 0xe8, 0x7f, 0x39, 0x4f, 0x79,
		0x9c, 0x00, 0xa5, 0x21, 0x03, 0xcb, 0x4e, 0x17,
	}
)

// Sign performs the schnorr adaptor signature generation
func Sign(privKey *btcec.PrivateKey, hash []byte, adaptorPointBytes []byte) (*Signature, error) {
	// The algorithm for producing a BIP-340 signature is described in
	// README.md and is reproduced here for reference:
	//
	// G = curve generator
	// n = curve order
	// d = private key
	// m = message
	// a = input randmoness
	// r, s = signature
	//
	// 1. d' = int(d)
	// 2. Fail if m is not 32 bytes
	// 3. Fail if d = 0 or d >= n
	// 4. P = d'*G
	// 5. Negate d if P.y is odd
	// 6. t = bytes(d) xor tagged_hash("BIP0340/aux", t || bytes(P) || m)
	// 7. rand = tagged_hash("BIP0340/nonce", a)
	// 8. k' = int(rand) mod n
	// 9. Fail if k' = 0
	// 10. R = 'k*G
	// 11. Negate k if R.y id odd
	// 12. e = tagged_hash("BIP0340/challenge", bytes(R) || bytes(P) || mod) mod n
	// 13. sig = bytes(R) || bytes((k + e*d)) mod n
	// 14. If Verify(bytes(P), m, sig) fails, abort.
	// 15. return sig.
	//
	// Note that the set of functional options passed in may modify the
	// above algorithm. Namely if CustomNonce is used, then steps 6-8 are
	// replaced with a process that generates the nonce using rfc6679. If
	// FastSign is passed, then we skip set 14.

	// Step 1.
	//
	// d' = int(d)
	var privKeyScalar btcec.ModNScalar
	privKeyScalar.Set(&privKey.Key)

	// Step 2.
	//
	// Fail if m is not 32 bytes
	if len(hash) != scalarSize {
		str := fmt.Sprintf("wrong size for message hash (got %v, want %v)",
			len(hash), scalarSize)
		return nil, fmt.Errorf("invalid hash length: %s", str)
	}

	// Step 3.
	//
	// Fail if d = 0 or d >= n
	if privKeyScalar.IsZero() {
		str := "private key is zero"
		return nil, fmt.Errorf("invalid private key", str)
	}

	// Step 4.
	//
	// P = 'd*G
	pub := privKey.PubKey()

	// Step 5.
	//
	// Negate d if P.y is odd.
	pubKeyBytes := pub.SerializeCompressed()
	if pubKeyBytes[0] == secp.PubKeyFormatCompressedOdd {
		privKeyScalar.Negate()
	}

	adaptorPoint, err := btcec.ParsePubKey(adaptorPointBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid adaptor point")
	}

	var privKeyBytes [scalarSize]byte
	privKeyScalar.PutBytes(&privKeyBytes)
	defer zeroArray(&privKeyBytes)
	for iteration := uint32(0); ; iteration++ {
		// Step 6-9.
		//
		// Use RFC6979 to generate a deterministic nonce k in [1, n-1]
		// parameterized by the private key, message being signed, extra data
		// that identifies the scheme, and an iteration count
		k := btcec.NonceRFC6979(
			privKeyBytes[:], hash, rfc6979ExtraDataV0[:], nil, iteration,
		)

		// Steps 10-15.
		sig, err := schnorrAdaptorSign(&privKeyScalar, k, pub, hash, adaptorPoint)
		k.Zero()
		if err != nil {
			// Try again with a new nonce.
			continue
		}

		return sig, nil
	}
}

func schnorrAdaptorSign(privKey, nonce *btcec.ModNScalar, pubKey *btcec.PublicKey, hash []byte,
	adaptorPoint *btcec.PublicKey) (*Signature, error) {

	// The algorithm for producing a BIP-340 signature is described in
	// README.md and is reproduced here for reference:
	//
	// G = curve generator
	// n = curve order
	// d = private key
	// m = message
	// a = input randmoness
	// r, s = signature
	//
	// 1. d' = int(d)
	// 2. Fail if m is not 32 bytes
	// 3. Fail if d = 0 or d >= n
	// 4. P = d'*G
	// 5. Negate d if P.y is odd
	// 6. t = bytes(d) xor tagged_hash("BIP0340/aux", t || bytes(P) || m)
	// 7. rand = tagged_hash("BIP0340/nonce", a)
	// 8. k' = int(rand) mod n
	// 9. Fail if k' = 0
	// 10. R = 'k*G
	// 11. Negate k if R.y id odd
	// 12. e = tagged_hash("BIP0340/challenge", bytes(R) || bytes(P) || m) mod n
	// 13. sig = bytes(R) || bytes((k + e*d)) mod n
	// 14. If Verify(bytes(P), m, sig) fails, abort.
	// 15. return sig.
	//
	// Note that the set of functional options passed in may modify the
	// above algorithm. Namely if CustomNonce is used, then steps 6-8 are
	// replaced with a process that generates the nonce using rfc6679. If
	// FastSign is passed, then we skip set 14.

	// NOTE: Steps 1-9 are performed by the caller.

	//
	// Step 10.
	//
	// R = kG
	var R btcec.JacobianPoint
	k := *nonce
	btcec.ScalarBaseMultNonConst(&k, &R)

	var AP, AR btcec.JacobianPoint
	adaptorPoint.AsJacobian(&AP)
	btcec.AddNonConst(&R, &AP, &AR)

	// Step 11.
	//
	// Negate nonce k if R.y is odd (R.y is the y coordinate of the point R)
	//
	// Note that R must be in affine coordinates for this check.
	AR.ToAffine()
	if AR.Y.IsOdd() {
		k.Negate()
	}

	// Step 12.
	//
	// e = tagged_hash("BIP0340/challenge", bytes(R) || bytes(P) || m) mod n
	var rBytes [32]byte
	r := &AR.X
	r.PutBytesUnchecked(rBytes[:])
	pBytes := schnorr.SerializePubKey(pubKey)

	commitment := chainhash.TaggedHash(
		chainhash.TagBIP0340Challenge, rBytes[:], pBytes, hash,
	)

	var e btcec.ModNScalar
	if overflow := e.SetBytes((*[32]byte)(commitment)); overflow != 0 {
		k.Zero()
		str := "hash of (r || P || m) too big"
		return nil, fmt.Errorf("incorrect schnorr hash: %s", str)
	}

	// Step 13.
	//
	// s = k + e*d mod n
	s := new(btcec.ModNScalar).Mul2(&e, privKey).Add(&k)
	k.Zero()

	R.ToAffine()

	sig := &Signature{
		r: R.X,
		s: *s,
	}

	// Step 14.
	//
	// If Verify(bytes(P), m, sig) fails, abort.
	// if !opts.fastSign {
	// 	if err := schnorrVerify(sig, hash, pBytes); err != nil {
	// 		return nil, err
	// 	}
	// }

	// Step 15.
	//
	// Return (r, s)
	return sig, nil
}

// zeroArray zeroes the memory of a scalar array.
func zeroArray(a *[scalarSize]byte) {
	for i := 0; i < scalarSize; i++ {
		a[i] = 0x00
	}
}
