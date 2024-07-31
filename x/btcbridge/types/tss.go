package types

import (
	"encoding/hex"
	"reflect"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ParticipantExists returns true if the given address is a participant, false otherwise
func ParticipantExists(participants []*DKGParticipant, addr string) bool {
	for _, p := range participants {
		if p.Address == addr {
			return true
		}
	}

	return false
}

// CheckCompletionRequests checks if the vaults of all the completion requests are same
func CheckCompletionRequests(requests []*DKGCompletionRequest) bool {
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

// GetVaultAddressFromPubKey gets the vault address from the given public key
// Note: the method generates taproot address
func GetVaultAddressFromPubKey(pubKey string) (string, error) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return "", err
	}

	parsedPubKey, err := schnorr.ParsePubKey(pubKeyBytes)
	if err != nil {
		return "", err
	}

	address, err := GetTaprootAddress(parsedPubKey, sdk.GetConfig().GetBtcChainCfg())
	if err != nil {
		return "", err
	}

	return address.String(), nil
}
