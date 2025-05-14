package types

// TSS module event types and attribute keys
const (
	EventTypeInitiateDKG         = "initiate_dkg"
	EventTypeCompleteDKG         = "complete_dkg"
	EventTypeInitiateSigning     = "initiate_signing"
	EventTypeCompleteSigning     = "complete_signing"
	EventTypeInitiateRefreshing  = "initiate_refreshing"
	EventTypeCompleteRefreshing  = "complete_refreshing"
	EventTypeRefreshingCompleted = "refreshing_completed"

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

	AttributeKeyDKGId               = "dkg_id"
	AttributeKeyRemovedParticipants = "removed_participants"
)

const (
	AttributeValueSeparator = ","
)
