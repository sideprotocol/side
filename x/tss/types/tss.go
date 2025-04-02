package types

import (
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
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

// DKGRequestCompletedHandler defines the callback handler on the DKG request completed
type DKGRequestCompletedHandler func(id uint64, ty string, intent int32, pubKeys []string) error

// ParticipantExists returns true if the given participant is included in the authorized participants, false otherwise
func ParticipantExists(participants []string, participant string) bool {
	return slices.Contains(participants, participant)
}

// CheckDKGCompletions checks if public keys of all DKG completions are same
func CheckDKGCompletions(completions []*DKGCompletion) bool {
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

// VerifySignature verifies the ed25519 signature against the given pub key and msg
// Assume that the signature and pub key are hex encoded
func VerifySignature(signature string, pubKey string, msg []byte) bool {
	sigBytes, _ := hex.DecodeString(signature)
	pubKeyBytes, _ := hex.DecodeString(pubKey)

	return ed25519.Verify(pubKeyBytes, msg, sigBytes)
}

// GetSigMsg gets the msg to be signed from the given data
// Assume that the given pub keys are hex encoded
func GetSigMsg(id uint64, pubKeys []string) []byte {
	rawMsg := make([]byte, 8)
	binary.BigEndian.PutUint64(rawMsg, id)

	for _, pubKey := range pubKeys {
		pubKeyBytes, _ := hex.DecodeString(pubKey)
		rawMsg = append(rawMsg, pubKeyBytes...)
	}

	return hash.Sha256(rawMsg)
}
