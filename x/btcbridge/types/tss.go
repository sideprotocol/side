package types

import (
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"reflect"

	"github.com/cometbft/cometbft/crypto"
)

// ParticipantExists returns true if the given address is a participant, false otherwise
func ParticipantExists(participants []*DKGParticipant, consAddress string) bool {
	for _, p := range participants {
		if p.ConsensusAddress == consAddress {
			return true
		}
	}

	return false
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

// VerifySignature verifies the given signature against the given DKG completion request
func VerifySignature(signature string, pubKey []byte, req *DKGCompletionRequest) bool {
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	sigMsg := GetSigMsgFromDKGCompletionReq(req)

	return ed25519.Verify(pubKey, sigMsg, sig)
}

// GetSigMsgFromDKGCompletionReq gets the msg to be signed from the given DKG completion request
func GetSigMsgFromDKGCompletionReq(req *DKGCompletionRequest) []byte {
	rawMsg := make([]byte, 8)
	binary.BigEndian.PutUint64(rawMsg, req.Id)

	for _, v := range req.Vaults {
		rawMsg = append(rawMsg, []byte(v)...)
	}

	return crypto.Sha256(rawMsg)
}
