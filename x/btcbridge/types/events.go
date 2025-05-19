package types

// BtcBridge module event types and attribute keys
const (
	EventTypeInitiateDKG      = "initiate_dkg_bridge"
	EventTypeInitiateSigning  = "initiate_signing_bridge"
	EventTypeIBCTransfer      = "ibc_transfer_bridge"
	EventTypeIBCWithdrawQueue = "ibc_withdraw_queue"
	EventTypeIBCWithdraw      = "ibc_withdraw"

	AttributeKeyId = "id"

	AttributeKeyParticipants   = "participants"
	AttributeKeyThreshold      = "threshold"
	AttributeKeyBatchSize      = "batch_size"
	AttributeKeyExpirationTime = "expiration_time"

	AttributeKeySigners   = "signers"
	AttributeKeySigHashes = "sig_hashes"

	AttributeKeyPacketSequence = "packet_sequence"
	AttributeKeyErrorMsg       = "err_msg"

	AttributeKeyAddress   = "address"
	AttributeKeyAmount    = "amount"
	AttributeKeySequence  = "sequence"
	AttributeKeyChannelId = "channel_id"
)

const (
	AttributeValueSeparator = ","
)
