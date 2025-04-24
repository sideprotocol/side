package types

// BtcBridge module event types and attribute keys
const (
	EventTypeInitiateDKG     = "initiate_dkg"
	EventTypeInitiateSigning = "initiate_signing"

	AttributeKeyId = "id"

	AttributeKeyParticipants   = "participants"
	AttributeKeyThreshold      = "threshold"
	AttributeKeyBatchSize      = "batch_size"
	AttributeKeyExpirationTime = "expiration_time"

	AttributeKeySigners   = "signers"
	AttributeKeySigHashes = "sig_hashes"
)

const (
	AttributeValueSeparator = ","
)
