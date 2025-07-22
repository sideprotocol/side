package types

// DLC module event types
const (
	EventTypeCreateOracle    = "create_oracle"
	EventTypeCreateDCM       = "create_dcm"
	EventTypeTriggerDLCEvent = "trigger_dlc_event"

	AttributeKeyId           = "id"
	AttributeKeyDKGId        = "dkg_id"
	AttributeKeyPubKey       = "pub_key"
	AttributeKeyDLCEventType = "dlc_event_type"
	AttributeKeyOutcome      = "outcome"
)

const (
	AttributeValueSeparator = ","
)
