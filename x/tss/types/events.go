package types

// TSS module event types
const (
	EventTypeInitiate = "initiate_dkg"
	EventTypeSign     = "sign"

	AttributeKeyId = "id"

	AttributeKeyParticipants   = "participants"
	AttributeKeyThreshold      = "threshold"
	AttributeKeyExpirationTime = "expiration_time"

	AttributeKeyType         = "type"
	AttributeKeyPubKey       = "pub_key"
	AttributeKeyNonce        = "nonce"
	AttributeKeyAdaptorPoint = "adaptor_point"
	AttributeKeySigHashes    = "sig_hashes"
)

const (
	AttributeValueSeparator = ","
)
