package types

// DLC module event types
const (
	EventTypeCreateOracle      = "create_oracle"
	EventTypeCreateAgency      = "create_agency"
	EventTypeGenerateNonce     = "generate_nonce"
	EventTypeTriggerPriceEvent = "trigger_price_event"

	AttributeKeyId             = "id"
	AttributeKeyParticipants   = "participants"
	AttributeKeyThreshold      = "threshold"
	AttributeKeyExpirationTime = "expiration_time"

	AttributeKeyEventId = "event_id"
	AttributeKeyPubKey  = "pub_key"
	AttributeKeyNonce   = "nonce"
	AttributeKeyPrice   = "price"

	AttributeKeyOraclePubKey = "oracle_pub_key"
)

const (
	AttributeValueSeparator = ","
)
