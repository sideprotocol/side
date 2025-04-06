package types

// DLC module event types
const (
	EventTypeCreateOracle    = "create_oracle"
	EventTypeCreateDCM       = "create_dcm"
	EventTypeGenerateNonce   = "generate_nonces"
	EventTypeTriggerDLCEvent = "trigger_dlc_event"

	AttributeKeyPubKey = "pub_key"

	AttributeKeyId             = "id"
	AttributeKeyDLCEventType   = "dlc_event_type"
	AttributeKeyParticipants   = "participants"
	AttributeKeyThreshold      = "threshold"
	AttributeKeyExpirationTime = "expiration_time"

	AttributeKeyOutcome = "outcome"

	AttributeKeyOraclePubKey = "oracle_pub_key"
)

const (
	AttributeValueSeparator = ","
)
