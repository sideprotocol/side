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

	TotalRewardsKey = []byte{0x10} // key for total distributed rewards
	RewardKeyPrefix = []byte{0x11} // prefix for each key to a reward
)

func RewardKey(address string) []byte {
	return append(RewardKeyPrefix, []byte(address)...)
}
