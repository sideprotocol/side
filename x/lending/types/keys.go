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
	// RepaymentEscrowAccount defines a escrow account for repayment
	RepaymentEscrowAccount = ModuleName + "_escrow"
)

var (
	Percent  = math.NewInt(100)
	Permille = math.NewInt(1000)

	ParamsKey = []byte{0x01}
	PriceKey  = []byte{0x02}

	PoolKeyPrefix         = []byte{0x10}
	LoanKeyPrefix         = []byte{0x11}
	DepositLogKeyPrefix   = []byte{0x12}
	RepaymentKeyPrefix    = []byte{0x13}
	DLCMetaKeyPrefix      = []byte{0x14}
	CancellationKeyPrefix = []byte{0x15}

	LoanByAddressKeyPrefix = []byte{0x16}
)

func PoolKey(id string) []byte {
	return append(PoolKeyPrefix, []byte(id)...)
}

func LoanKey(id string) []byte {
	return append(LoanKeyPrefix, []byte(id)...)
}

func LoanByAddressKey(id string, address string) []byte {
	return append(append(LoanByAddressKeyPrefix, []byte(address)...), []byte(id)...)
}

func DepositLogKey(txid string) []byte {
	return append(DepositLogKeyPrefix, []byte(txid)...)
}

func DLCMetaKey(loanId string) []byte {
	return append(DLCMetaKeyPrefix, []byte(loanId)...)
}

func RepaymentKey(loanId string) []byte {
	return append(RepaymentKeyPrefix, []byte(loanId)...)
}

func CancellationKey(loanId string) []byte {
	return append(CancellationKeyPrefix, []byte(loanId)...)
}
