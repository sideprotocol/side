package types

// Farming module event types and attribute keys
const (
	EventTypeStake   = "stake"
	EventTypeUnstake = "unstake"
	EventTypeClaim   = "claim"

	AttributeKeyStaker       = "staker"
	AttributeKeyId           = "id"
	AttributeKeyAmount       = "amount"
	AttributeKeyLockDuration = "lock_duration"
	AttributeKeyRewards      = "rewards"
)
