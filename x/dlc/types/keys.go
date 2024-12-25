package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "dlc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_dlc"
)

var (
	ParamsKey        = []byte{0x01} // key for params
	OracleIdKey      = []byte{0x02} // key for oracle id
	AgencyIdKey      = []byte{0x03} // key for agency id
	EventIdKey       = []byte{0x04} // key for event id
	AttestationIdKey = []byte{0x05} // key for attestation id

	OracleKeyPrefix         = []byte{0x10} // prefix for each key to an oracle
	OracleByPubKeyKeyPrefix = []byte{0x11} // prefix for each key to an oracle by public key
	AgencyKeyPrefix         = []byte{0x12} // prefix for each key to an agency
	NonceIndexKeyPrefix     = []byte{0x13} // key prefix for the nonce index
	NonceKeyPrefix          = []byte{0x14} // prefix for each key to a nonce
	EventKeyPrefix          = []byte{0x15} // prefix for each key to an event
	AttestationKeyPrefix    = []byte{0x16} // prefix for each key to an attestation

	PriceKeyPrefix = []byte{0x20} // key prefix for the price
)

func OracleKey(id uint64) []byte {
	return append(OracleKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func OracleByPubKeyKey(pubKey []byte) []byte {
	return append(OracleByPubKeyKeyPrefix, pubKey...)
}

func AgencyKey(id uint64) []byte {
	return append(AgencyKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func NonceIndexKey(oracleId uint64) []byte {
	return append(NonceIndexKeyPrefix, sdk.Uint64ToBigEndian(oracleId)...)
}

func NonceKey(oracleId uint64, index uint64) []byte {
	return append(append(NonceKeyPrefix, sdk.Uint64ToBigEndian(oracleId)...), sdk.Uint64ToBigEndian(index)...)
}

func EventKey(id uint64) []byte {
	return append(EventKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func AttestationKey(id uint64) []byte {
	return append(AttestationKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func PriceKey(pair string) []byte {
	return append(PriceKeyPrefix, []byte(pair)...)
}
