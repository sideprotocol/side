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
	EpochIdKey   = []byte{0x02} // key for epoch id
	StakingIdKey = []byte{0x03} // key for staking id

	EpochKeyPrefix            = []byte{0x10} // key prefix for epoch
	StakingKeyPrefix          = []byte{0x11} // key prefix for staking
	StakingByStatusKeyPrefix  = []byte{0x12} // key prefix for staking by status
	StakingByAddressKeyPrefix = []byte{0x13} // key prefix for staking by address
	TotalStakingKeyPrefix     = []byte{0x14} // key prefix for total staking

	CurrentEpochStakingQueueKeyPrefix          = []byte{0x20} // key prefix for staking queue for the current epoch
	CurrentEpochStakingQueueByAddressKeyPrefix = []byte{0x21} // key prefix for staking queue by address for the current epoch
)

func EpochKey(id uint64) []byte {
	return append(EpochKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func StakingKey(id uint64) []byte {
	return append(StakingKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func StakingByStatusKey(status StakingStatus, id uint64) []byte {
	key := append(StakingByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...)

	return append(key, sdk.Uint64ToBigEndian(id)...)
}

func StakingByAddressKey(address string, id uint64) []byte {
	return append(append(StakingByAddressKeyPrefix, []byte(address)...), sdk.Uint64ToBigEndian(id)...)
}

func TotalStakingKey(denom string) []byte {
	return append(TotalStakingKeyPrefix, []byte(denom)...)
}

func CurrentEpochStakingQueueKey(stakingId uint64) []byte {
	return append(CurrentEpochStakingQueueKeyPrefix, sdk.Uint64ToBigEndian(stakingId)...)
}

func CurrentEpochStakingQueueByAddressKey(address string, stakingId uint64) []byte {
	return append(append(CurrentEpochStakingQueueByAddressKeyPrefix, []byte(address)...), sdk.Uint64ToBigEndian(stakingId)...)
}
