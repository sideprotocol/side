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
)
