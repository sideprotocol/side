package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "liquidation"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_liquidation"
)

var (
	ParamsKey              = []byte{0x01} // key for params
	LiquidationIdKey       = []byte{0x02} // key for liquidation id
	LiquidationRecordIdKey = []byte{0x03} // key for liquidation record id

	LiquidationKeyPrefix                    = []byte{0x10} // prefix for each key to a liquidation
	LiquidationByStatusKeyPrefix            = []byte{0x11} // prefix for each key to a liquidation by status
	LiquidationRecordKeyPrefix              = []byte{0x12} // prefix for each key to a liquidation record
	LiquidationRecordByLiquidationKeyPrefix = []byte{0x13} // prefix for each key to a liquidation record by liquidation
)

func LiquidationKey(id uint64) []byte {
	return append(LiquidationKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func LiquidationByStatusKey(status LiquidationStatus, id uint64) []byte {
	key := append(LiquidationByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...)

	return append(key, sdk.Uint64ToBigEndian(id)...)
}

func LiquidationRecordKey(id uint64) []byte {
	return append(LiquidationRecordKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func LiquidationRecordByLiquidationKey(liquidationId uint64, recordId uint64) []byte {
	key := append(LiquidationRecordByLiquidationKeyPrefix, sdk.Uint64ToBigEndian(liquidationId)...)

	return append(key, sdk.Uint64ToBigEndian(recordId)...)
}
