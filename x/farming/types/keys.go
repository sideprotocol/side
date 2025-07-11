package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "farming"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_farming"
)

var (
	ParamsKey    = []byte{0x01} // key for params
	StakingIdKey = []byte{0x02} // key for staking id

	StakingKeyPrefix          = []byte{0x10} // key prefix for staking
	StakingByAddressKeyPrefix = []byte{0x11} // key prefix for staking by address
	TotalStakingsKeyPrefix    = []byte{0x12} // key prefix for total stakings
)

func StakingKey(id uint64) []byte {
	return append(StakingKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func StakingByAddressKey(address string, id uint64) []byte {
	return append(append(StakingByAddressKeyPrefix, []byte(address)...), sdk.Uint64ToBigEndian(id)...)
}

func TotalStakingsKey(denom string) []byte {
	return append(TotalStakingsKeyPrefix, []byte(denom)...)
}
