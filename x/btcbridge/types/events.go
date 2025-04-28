package types

// BtcBridge module event types and attribute keys
const (
	EventTypeInitiateDKG     = "initiate_dkg_bridge"
	EventTypeInitiateSigning = "initiate_signing_bridge"

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
