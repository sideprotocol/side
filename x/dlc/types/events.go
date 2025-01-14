package types

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
