package types

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"reflect"
	time "time"

	"github.com/sideprotocol/side/bitcoin/crypto/hash"
)

// ParticipantExists returns true if the given participant is included in the authorized participants, false otherwise
func ParticipantExists(participants []*DKGParticipant, consPubKey string) bool {
	for _, p := range participants {
		if p.ConsensusPubkey == consPubKey {
			return true
		}
	}

	return false
}

// GetParticipantPubKeys gets consensus pub keys of all participants
func GetParticipantPubKeys(participants []*DKGParticipant) []string {
	pubKeys := []string{}

	for _, p := range participants {
		pubKeys = append(pubKeys, p.ConsensusPubkey)
	}

	return pubKeys
}

// CheckDKGCompletionRequests checks if the vaults of all the DKG completion requests are same
func CheckDKGCompletionRequests(requests []*DKGCompletionRequest) bool {
	if len(requests) == 0 {
		return false
	}

	vaults := requests[0].Vaults

	for _, req := range requests[1:] {
		if !reflect.DeepEqual(req.Vaults, vaults) {
			return false
		}
	}

	return true
}

// VerifySignature verifies the ed25519 signature against the given pub key and msg
// Assume that the signature is hex encoded and the pub key is base64 encoded
func VerifySignature(signature string, pubKey string, msg []byte) bool {
	sigBytes, _ := hex.DecodeString(signature)
	pubKeyBytes, _ := base64.StdEncoding.DecodeString(pubKey)

	return ed25519.Verify(pubKeyBytes, msg, sigBytes)
}

// GetDKGCompletionSigMsg gets the msg to be signed from the given DKG completion request
func GetDKGCompletionSigMsg(req *DKGCompletionRequest) []byte {
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, req.Id)

	for _, v := range req.Vaults {
		msg = append(msg, []byte(v)...)
	}

	return hash.Sha256(msg)
}

// GetRefreshingCompletionSigMsg gets the msg to be signed from the given data for the refreshing completion
func GetRefreshingCompletionSigMsg(id uint64) []byte {
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, id)

	return hash.Sha256(msg)
}

// GetExpirationTime gets the expiration time according to the given timeout duration
func GetExpirationTime(currentTime time.Time, timeoutDuration time.Duration) time.Time {
	if timeoutDuration == 0 {
		return time.Time{}
	}

	return currentTime.Add(timeoutDuration)
}
