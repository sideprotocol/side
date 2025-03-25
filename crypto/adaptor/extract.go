package adaptor

import (
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Extract extracts the secret from the given adaptor signature and adapted signature
func Extract(adaptorSigBytes []byte, adaptedSigBytes []byte) []byte {
	adaptorSig, err := ParseSignature(adaptorSigBytes)
	if err != nil {
		return nil
	}

	_, err = schnorr.ParseSignature(adaptedSigBytes)
	if err != nil {
		return nil
	}

	var adaptedS secp256k1.ModNScalar
	adaptedS.SetByteSlice(adaptedSigBytes[32:])

	t := adaptedS.Add(adaptorSig.s.Negate())

	if adaptorSig.IsROdd() {
		t.Negate()
	}

	return SerializeScalar(t)
}
