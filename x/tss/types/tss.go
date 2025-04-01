package types

import (
	"crypto/ed25519"
	"encoding/binary"
	"reflect"
	"slices"

	"github.com/sideprotocol/side/crypto/hash"
)

const (
	// schnorr signature size
	SchnorrSignatureSize = 64

	// schnorr adaptor signature size
	SchnorrAdaptorSignatureSize = 65
)

// ParticipantExists returns true if the given participant is an authorized participant, false otherwise
func ParticipantExists(participants []string, participant string) bool {
	return slices.Contains(participants, participant)
}

// CheckDKGCompletions checks if public keys of all the DKG completion are same
func CheckDKGCompletionRequests(completions []*DKGCompletion) bool {
	if len(completions) == 0 {
		return false
	}

	pubKeys := completions[0].PubKeys

	for _, completion := range completions[1:] {
		if !reflect.DeepEqual(completion.PubKeys, pubKeys) {
			return false
		}
	}

	return true
}

// GetSigMsg gets the msg to be signed from the given data
func GetSigMsg(id uint64, pubKeys [][]byte) []byte {
	rawMsg := make([]byte, 8)
	binary.BigEndian.PutUint64(rawMsg, id)

	for _, pubKey := range pubKeys {
		rawMsg = append(rawMsg, pubKey...)
	}

	return hash.Sha256(rawMsg)
}

// VerifySignature verifies the given signature
func VerifySignature(signature []byte, pubKey []byte, msg []byte) bool {
	return ed25519.Verify(pubKey, msg, signature)
}
