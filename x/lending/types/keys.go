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
	Percent        = math.NewInt(100)
	Permille       = math.NewInt(1000)
	ParamsStoreKey = []byte{0x1}

	PoolStorePrefix  = []byte{0x2}
	LoanStorePrefix  = []byte{0x3}
	DepositLogPrefix = []byte{0x4}
	RepaymentPrefix  = []byte{0x5}
	LoanCETKeyPrefix = []byte{0x06}
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

func LoanCETKey(loanId string) []byte {
	return append(LoanCETKeyPrefix, []byte(loanId)...)
}

func RepaymentKey(loanId string) []byte {
	return append(RepaymentPrefix, []byte(loanId)...)
}
