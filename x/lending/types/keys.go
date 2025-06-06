package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	ParamsKey       = []byte{0x01}
	PriceKeyPrefix  = []byte{0x02}
	RedemptionIdKey = []byte{0x03}

	PoolKeyPrefix            = []byte{0x10}
	LoanKeyPrefix            = []byte{0x11}
	LoanByAddressKeyPrefix   = []byte{0x12}
	AuthorizationIdKeyPrefix = []byte{0x13}
	DepositLogKeyPrefix      = []byte{0x14}
	RepaymentKeyPrefix       = []byte{0x15}
	DLCMetaKeyPrefix         = []byte{0x16}
	RedemptionKeyPrefix      = []byte{0x17}
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

func AuthorizationIdKey(loanId string) []byte {
	return append(AuthorizationIdKeyPrefix, []byte(loanId)...)
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

func RedemptionKey(id uint64) []byte {
	return append(RedemptionKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func PriceKey(pair string) []byte {
	return append(PriceKeyPrefix, []byte(pair)...)
}
