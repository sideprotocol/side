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
	PhaseIdKey   = []byte{0x03} // key for phase id

	PhaseKeyPrefix            = []byte{0x10} // key prefix for phase
	StakingKeyPrefix          = []byte{0x11} // key prefix for staking
	StakingByAddressKeyPrefix = []byte{0x12} // key prefix for staking by address
	TotalStakingKeyPrefix     = []byte{0x13} // key prefix for total staking
)

func PhaseKey(id uint64) []byte {
	return append(PhaseKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func StakingKey(id uint64) []byte {
	return append(StakingKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func StakingByAddressKey(address string, id uint64) []byte {
	return append(append(StakingByAddressKeyPrefix, []byte(address)...), sdk.Uint64ToBigEndian(id)...)
}

func TotalStakingKey(phaseId uint64, denom string) []byte {
	return append(append(TotalStakingKeyPrefix, sdk.Uint64ToBigEndian(phaseId)...), []byte(denom)...)
}
