package types

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"slices"

	"github.com/sideprotocol/side/crypto/hash"
)

// ParticipantExists returns true if the given public key is a participant, false otherwise
func ParticipantExists(participants []string, pubKey string) bool {
	return slices.Contains(participants, pubKey)
}

// CheckPendingPubKeys checks if all pending public keys are same
func CheckPendingPubKeys(pubKeys [][]byte) bool {
	if len(pubKeys) == 0 {
		return false
	}

	expectedPubKey := pubKeys[0]

	for _, pk := range pubKeys[1:] {
		if !bytes.Equal(pk, expectedPubKey) {
			return false
		}
	}

	return true
}

// GetSigMsg gets the msg to be signed from the given data
func GetSigMsg(id uint64, pubKey []byte) []byte {
	rawMsg := make([]byte, 8)
	binary.BigEndian.PutUint64(rawMsg, id)

	rawMsg = append(rawMsg, pubKey...)

	return hash.Sha256(rawMsg)
}

// VerifySignature verifies the given signature
func VerifySignature(signature []byte, pubKey []byte, msg []byte) bool {
	return ed25519.Verify(pubKey, msg, signature)
}
