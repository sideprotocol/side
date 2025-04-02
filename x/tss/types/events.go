package types

// TSS module event types
const (
	EventTypeInitiateDKG = "initiate_dkg"
	EventTypeCompleteDKG = "complete_dkg"
	EventTypeSign        = "sign"

	AttributeKeySender = "sender"

	AttributeKeyId     = "id"
	AttributeKeyModule = "module"
	AttributeKeyType   = "type"
	AttributeKeyIntent = "intent"

	AttributeKeyParticipants   = "participants"
	AttributeKeyThreshold      = "threshold"
	AttributeKeyBatchSize      = "batch_size"
	AttributeKeyExpirationTime = "expiration_time"

	AttributeKeyParticipant = "participant"

	AttributeKeyPubKey       = "pub_key"
	AttributeKeyNonce        = "nonce"
	AttributeKeyAdaptorPoint = "adaptor_point"
	AttributeKeySigHashes    = "sig_hashes"
)

const (
	AttributeValueSeparator = ","
)
