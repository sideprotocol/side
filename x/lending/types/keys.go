package types

import (
	"cosmossdk.io/math"
)

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
	Percent        = math.NewInt(100)
	Permille       = math.NewInt(1000)
	ParamsStoreKey = []byte{0x1}

	PoolStorePrefix  = []byte{0x2}
	LoanStorePrefix  = []byte{0x3}
	DepositLogPrefix = []byte{0x4}
)

func PoolStoreKey(pool_id string) []byte {
	return append(PoolStorePrefix, []byte(pool_id)...)
}

func LoanStoreKey(vault string) []byte {
	return append(LoanStorePrefix, []byte(vault)...)
}

func DepositLogKey(txid string) []byte {
	return append(DepositLogPrefix, []byte(txid)...)
}
