package types

const (
	// ModuleName defines the module name
	ModuleName = "lending"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_lending"
)

var (
	ParamsStoreKey = []byte{0x1}

	PoolStorePrefix = []byte{0x2}
	LoanStorePrefix = []byte{0x3}
)

func PoolStoreKey(pool_id string) []byte {
	return append(PoolStorePrefix, []byte(pool_id)...)
}

func LoanStoreKey(vault string) []byte {
	return append(LoanStorePrefix, []byte(vault)...)
}
