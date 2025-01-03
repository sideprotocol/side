package types

const (
	// ModuleName defines the module name
	ModuleName = "incentive"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_incentive"
)

var (
	ParamsStoreKey = []byte{0x1}

	RewardStatsKey   = []byte{0x11} // key for total reward statistics
	RewardsKeyPrefix = []byte{0x12} // prefix for each key to the rewards
)

func RewardsKey(address string) []byte {
	return append(RewardsKeyPrefix, []byte(address)...)
}
