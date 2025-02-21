package adaptor

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Verify verifies the provided schnorr adaptor signature against the given adaptor point
func Verify(sigBytes []byte, msg []byte, pubKeyBytes []byte, adaptorPointBytes []byte) bool {
	_, err := schnorr.ParseSignature(sigBytes)
	if err != nil {
		return false
	}

	pubKey, err := schnorr.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false
	}

	adaptorPoint, err := btcec.ParsePubKey(adaptorPointBytes)
	if err != nil {
		return false
	}

	return verifySchnorrAdaptorSignature(NewSignature(sigBytes), msg, pubKey, adaptorPoint) == nil
}

// verifySchnorrAdaptorSignature verifies the given schnorr adaptor signature
//
// The algorithm is based on the verifier for schnorr signature
// Specifically, step 3 and 6 are added, step 7 and 10 are modified
//
// Annotation:
// AP: adaptor point
// AR: adapted R
func verifySchnorrAdaptorSignature(sig *Signature, hash []byte, pubKey *secp256k1.PublicKey, adaptorPoint *secp256k1.PublicKey) error {
	// 1. Fail if m is not 32 bytes
	// 2. P = lift_x(int(pk)).
	// 3. AP = lift_x(int(ap)).
	// 4. r = int(sig[0:32]); fail is r >= p.
	// 5. s = int(sig[32:64]); fail if s >= n.
	// 6. AR = R + AP.
	// 7. e = int(tagged_hash("BIP0340/challenge", bytes(AR) || bytes(P) || M)) mod n.
	// 8. ER = s*G - e*P
	// 9. Fail if is_infinite(ER)
	// 10. Fail if not is_infinite(R+ER) in case not has_even_y(R+AP)
	// 11. Fail if x(ER) != r in case has_even_y(R+AP)
	// 12. Return success iff not failure occured before reaching this
	// point.

	// Step 1.
	//
	// Fail if m is not 32 bytes
	if len(hash) != scalarSize {
		str := fmt.Sprintf("wrong size for message (got %v, want %v)",
			len(hash), scalarSize)
		return fmt.Errorf("invalid hash length: %s", str)
	}

	// Step 2.
	//
	// P = lift_x(int(pk))
	//
	// Fail if P is not a point on the curve
	if !pubKey.IsOnCurve() {
		str := "pubkey point is not on curve"
		return fmt.Errorf("invalid pub key: %s", str)
	}

	// Step 3.
	//
	// AP = lift_x(int(ap))
	//
	// Fail if AP is not a point on the curve
	if !adaptorPoint.IsOnCurve() {
		str := "adaptor point is not on curve"
		return fmt.Errorf("invalid adaptor point: %s", str)
	}

	// Step 4.
	//
	// Fail if r >= p
	//
	// Note this is already handled by the fact r is a field element.

	// Step 5.
	//
	// Fail if s >= n
	//
	// Note this is already handled by the fact s is a mod n scalar.

	// Step 6.
	//
	// AR = R + AP
	var rBytes [32]byte
	sig.r.PutBytesUnchecked(rBytes[:])

	rPoint, err := schnorr.ParsePubKey(rBytes[:])
	if err != nil {
		str := "failed to parse r"
		return fmt.Errorf("invalid r: %s", str)
	}

	var R, AP, AR btcec.JacobianPoint
	rPoint.AsJacobian(&R)
	adaptorPoint.AsJacobian(&AP)
	btcec.AddNonConst(&R, &AP, &AR)

	// Step 7.
	//
	// e = int(tagged_hash("BIP0340/challenge", bytes(ar) || bytes(P) || M)) mod n.
	AR.ToAffine()
	var arBytes [32]byte
	AR.X.PutBytesUnchecked(arBytes[:])

	pBytes := schnorr.SerializePubKey(pubKey)

	commitment := chainhash.TaggedHash(
		chainhash.TagBIP0340Challenge, arBytes[:], pBytes, hash,
	)

	var e btcec.ModNScalar
	if overflow := e.SetBytes((*[32]byte)(commitment)); overflow != 0 {
		str := "hash of (r || P || m) too big"
		return fmt.Errorf("invalid schnorr hash: %s", str)
	}

	// Negate e here so we can use AddNonConst below to subtract the s*G
	// point from e*P.
	e.Negate()

	// Step 8.
	//
	// ER = s*G - e*P
	var P, sG, eP, ER btcec.JacobianPoint
	pubKey.AsJacobian(&P)
	btcec.ScalarBaseMultNonConst(&sig.s, &sG)
	btcec.ScalarMultNonConst(&e, &P, &eP)
	btcec.AddNonConst(&sG, &eP, &ER)

	// Step 9.
	//
	// Fail if ER is the point at infinity
	if (ER.X.IsZero() && ER.Y.IsZero()) || ER.Z.IsZero() {
		str := "calculated R point is the point at infinity"
		return fmt.Errorf("invalid signature: %s", str)
	}

	// Step 10.
	//
	// Fail if not is_infinite(R+ER) in case (R+AP).y is odd
	//
	// Note that R+AP must be in affine coordinates for this check.
	if AR.Y.IsOdd() {
		var Check btcec.JacobianPoint
		btcec.AddNonConst(&R, &ER, &Check)

		if !((Check.X.IsZero() && Check.Y.IsZero()) || Check.Z.IsZero()) {
			str := "effective R point is not negated R"
			return fmt.Errorf("invalid signature: %s", str)
		}
	} else {
		// Step 11.
		//
		// Verified if ER.x == r in case (R+AP).y is even
		//
		// Note that ER must be in affine coordinates for this check.
		ER.ToAffine()
		if ER.X != sig.r {
			str := "effective R point was not given R"
			return fmt.Errorf("invalid signature: %s", str)
		}
	}

	// Step 12.
	//
	// Return success iff not failure occured before reaching this
	return nil
}
