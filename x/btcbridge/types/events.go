package types

// BtcBridge module event types and attribute keys
const (
	EventTypeInitiateDKG            = "initiate_dkg_bridge"
	EventTypeInitiateSigning        = "initiate_signing_bridge"
	EventTypeInitiateRefreshing     = "initiate_refreshing_bridge"
	EventTypeCompleteRefreshing     = "complete_refreshing_bridge"
	EventTypeRefreshingCompleted    = "refreshing_completed_bridge"
	EventTypeIBCTransfer            = "ibc_transfer_bridge"
	EventTypeIBCWithdrawQueue       = "ibc_withdraw_queue"
	EventTypeIBCWithdraw            = "ibc_withdraw"
	EventTypeGlobalRateLimitUpdated = "global_rate_limit_updated"

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

	AttributeKeyDKGId               = "dkg_id"
	AttributeKeyRemovedParticipants = "removed_participants"

	AttributeKeySender      = "sender"
	AttributeKeyParticipant = "participant"

	AttributeKeyPreviousStartTime = "previous_start_time"
	AttributeKeyPreviousEndTime   = "previous_end_time"
	AttributeKeyPreviousQuota     = "previous_quota"
	AttributeKeyPreviousUsed      = "previous_used"
	AttributeKeyStartTime         = "start_time"
	AttributeKeyEndTime           = "end_time"
	AttributeKeyQuota             = "quota"
)

const (
	AttributeValueSeparator = ","
)
