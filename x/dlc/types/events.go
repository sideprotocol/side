package types

// DLC module event types
const (
	EventTypeCreateOracle    = "create_oracle"
	EventTypeCreateDCM       = "create_dcm"
	EventTypeTriggerDLCEvent = "trigger_dlc_event"

	AttributeKeyPubKey = "pub_key"

	AttributeKeyId           = "id"
	AttributeKeyDLCEventType = "dlc_event_type"
	AttributeKeyOutcome      = "outcome"
)

const (
	AttributeValueSeparator = ","
)
