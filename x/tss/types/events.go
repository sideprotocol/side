package types

// TSS module event types and attribute keys
const (
	EventTypeInitiateDKG       = "initiate_dkg"
	EventTypeCompleteDKG       = "complete_dkg"
	EventTypeInitiateSigning   = "initiate_signing"
	EventTypeCompleteSigning   = "complete_signing"
	EventTypeInitiateResharing = "initiate_resharing"
	EventTypeCompleteResharing = "complete_resharing"

	AttributeKeySender = "sender"

	AttributeKeyId       = "id"
	AttributeKeyModule   = "module"
	AttributeKeyScopedId = "scoped_id"
	AttributeKeyType     = "type"
	AttributeKeyIntent   = "intent"

	AttributeKeyParticipants   = "participants"
	AttributeKeyThreshold      = "threshold"
	AttributeKeyBatchSize      = "batch_size"
	AttributeKeyExpirationTime = "expiration_time"

	AttributeKeyParticipant = "participant"

	AttributeKeyPubKey    = "pub_key"
	AttributeKeySigHashes = "sig_hashes"
	AttributeKeyOption    = "option"

	AttributeKeyRemovedParticipants = "removed_participants"
	AttributeKeyNewParticipants     = "new_participants"
)

const (
	AttributeValueSeparator = ","
)
