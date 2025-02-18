package types

import (
	"encoding/hex"
	fmt "fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/sideprotocol/side/crypto/hash"
)

// GetSignaturePointFromEvent gets the signature point from the given event
func GetSignaturePointFromEvent(event *DLCPriceEvent) ([]byte, error) {
	oralcePubKey, err := hex.DecodeString(event.Pubkey)
	if err != nil {
		return nil, err
	}

	nonce, err := hex.DecodeString(event.Nonce)
	if err != nil {
		return nil, err
	}

	return GetSignaturePoint(oralcePubKey, nonce, hash.Sha256(event.TriggerPrice.BigInt().Bytes()))
}

// GetSignaturePoint gets the signature point from the given params
func GetSignaturePoint(pubKeyBytes []byte, nonceBytes []byte, msg []byte) ([]byte, error) {
	pubKey, err := schnorr.ParsePubKey(pubKeyBytes)
	if err != nil {
		return nil, err
	}

	nonce, err := schnorr.ParsePubKey(nonceBytes)
	if err != nil {
		return nil, err
	}

	commitment := chainhash.TaggedHash(
		chainhash.TagBIP0340Challenge, nonceBytes, pubKeyBytes, msg,
	)

	var e btcec.ModNScalar
	if overflow := e.SetBytes((*[32]byte)(commitment)); overflow != 0 {
		return nil, fmt.Errorf("invalid schnorr hash")
	}

	var P, R, eP, sG btcec.JacobianPoint
	pubKey.AsJacobian(&P)
	nonce.AsJacobian(&R)
	btcec.ScalarMultNonConst(&e, &P, &eP)
	btcec.AddNonConst(&R, &eP, &sG)

	return btcec.JacobianToByteSlice(sG), nil
}
