package types

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/sideprotocol/side/bitcoin/crypto/hash"
)

const (
	// DKG type for DCM creation
	DKG_TYPE_DCM = "dcm"

	// DKG type for nonce generation along with oracle
	DKG_TYPE_NONCE = "nonce"

	// default outcome index
	DefaultOutcomeIndex = -1
)

// GetEventOutcomeHash gets the event outcome hash by the given index
// Assume that the outcome index is valid
func GetEventOutcomeHash(event *DLCEvent, outcomeIndex int) []byte {
	return hash.Sha256([]byte(event.Outcomes[outcomeIndex]))
}

// GetSignaturePointFromEvent gets the signature point from the given event and outcome index
// Assume that the outcome index is valid
func GetSignaturePointFromEvent(event *DLCEvent, outcomeIndex int) ([]byte, error) {
	oraclePubKey, err := hex.DecodeString(event.Pubkey)
	if err != nil {
		return nil, err
	}

	nonce, err := hex.DecodeString(event.Nonce)
	if err != nil {
		return nil, err
	}

	return GetSignaturePoint(oraclePubKey, nonce, GetEventOutcomeHash(event, outcomeIndex))
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

// GetEventTypeFromIntent gets the event type from the given nonce DKG intent
func GetEventTypeFromIntent(intent int32) DlcEventType {
	switch intent {
	case int32(DKGIntent_DKG_INTENT_DATE_EVENT_NONCE):
		return DlcEventType_DATE

	case int32(DKGIntent_DKG_INTENT_LENDING_EVENT_NONCE):
		return DlcEventType_LENDING

	default:
		return DlcEventType_PRICE
	}
}

// EventStatusToByte converts the given status to a byte
func EventStatusToByte(triggered bool) byte {
	if triggered {
		return 1
	}

	return 0
}

// ByteToEventStatus converts the given byte to the status
func ByteToEventStatus(b byte) bool {
	return b != 0
}

// ToScopedId converts the given local id to the scoped id
func ToScopedId(id uint64) string {
	return fmt.Sprintf("%d", id)
}

// FromScopedId converts the scoped id to the local id
// Assume that the scoped id is valid
func FromScopedId(scopedId string) uint64 {
	id, _ := strconv.ParseUint(scopedId, 10, 64)
	return id
}
