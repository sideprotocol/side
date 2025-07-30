package types

import (
	"encoding/hex"
	"strings"

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
	ParamsKey       = []byte{0x01}
	RedemptionIdKey = []byte{0x02}

	PoolKeyPrefix             = []byte{0x10}
	LoanKeyPrefix             = []byte{0x11}
	LoanByStatusKeyPrefix     = []byte{0x12}
	LoanByAddressKeyPrefix    = []byte{0x13}
	LoanByOracleKeyPrefix     = []byte{0x14}
	LiquidationQueueKeyPrefix = []byte{0x15}
	AuthorizationIdKeyPrefix  = []byte{0x16}
	DepositLogKeyPrefix       = []byte{0x17}
	RepaymentKeyPrefix        = []byte{0x18}
	DLCMetaKeyPrefix          = []byte{0x19}
	RedemptionKeyPrefix       = []byte{0x20}

	ReferrerKeyPrefix = []byte{0x30}
)

func PoolKey(id string) []byte {
	return append(PoolKeyPrefix, []byte(id)...)
}

func LoanKey(id string) []byte {
	return append(LoanKeyPrefix, []byte(id)...)
}

func LoanByStatusKey(status LoanStatus, id string) []byte {
	key := append(LoanByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...)

	return append(key, []byte(id)...)
}

func LoanByAddressKey(id string, address string) []byte {
	return append(append(LoanByAddressKeyPrefix, []byte(address)...), []byte(id)...)
}

func LoanByOracleKey(oraclePubKey string, id string) []byte {
	pubKey, _ := hex.DecodeString(oraclePubKey)

	key := append(LoanByOracleKeyPrefix, pubKey...)

	return append(key, []byte(id)...)
}

func LiquidationQueueKey(loanId string) []byte {
	return append(LiquidationQueueKeyPrefix, []byte(loanId)...)
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

func ReferrerKey(referralCode string) []byte {
	return append(ReferrerKeyPrefix, []byte(strings.ToLower(referralCode))...)
}
