package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DLC module event types
const (
	EventTypeCreateOracle = "create_oracle"
	EventTypeCreateAgency = "create_agency"
	EventTypeTriggerEvent = "trigger_event"

	AttributeKeyId             = "id"
	AttributeKeyParticipants   = "participants"
	AttributeKeyThreshold      = "threshold"
	AttributeKeyExpirationTime = "expiration_time"

	AttributeKeyEventId = "event_id"
	AttributeKeyPubKey  = "pub_key"
	AttributeKeyNonce   = "nonce"
	AttributeKeyPrice   = "price"
)

// GetParticipantsAttributes gets the attribute list for the given participants
func GetParticipantsAttributes(participants []string) []sdk.Attribute {
	attributes := []sdk.Attribute{}

	for _, p := range participants {
		attributes = append(attributes, sdk.NewAttribute(AttributeKeyParticipants, p))
	}

	return attributes
}
