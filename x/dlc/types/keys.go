package types

import (
	sdkmath "cosmossdk.io/math"
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

	OracleKeyPrefix              = []byte{0x10} // prefix for each key to an oracle
	OracleByPubKeyKeyPrefix      = []byte{0x11} // prefix for each key to an oracle by public key
	PendingOraclePubKeyKeyPrefix = []byte{0x12} // key prefix for the pending oracle public key
	AgencyKeyPrefix              = []byte{0x13} // prefix for each key to an agency
	PendingAgencyPubKeyKeyPrefix = []byte{0x14} // key prefix for the pending agency public key
	NonceIndexKeyPrefix          = []byte{0x15} // key prefix for the nonce index
	NonceKeyPrefix               = []byte{0x16} // prefix for each key to a nonce
	EventKeyPrefix               = []byte{0x17} // prefix for each key to an event
	EventByPriceKeyPrefix        = []byte{0x18} // prefix for each key to an event by triggering price
	CurrentEventPriceKeyPrefix   = []byte{0x19} // key prefix for the current event price
	AttestationKeyPrefix         = []byte{0x20} // prefix for each key to an attestation
	AttestationByEventKeyPrefix  = []byte{0x21} // prefix for each key to an attestation by event

	PriceKeyPrefix = []byte{0x30} // key prefix for the price
)

func OracleKey(id uint64) []byte {
	return append(OracleKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func OracleByPubKeyKey(pubKey []byte) []byte {
	return append(OracleByPubKeyKeyPrefix, pubKey...)
}

func PendingOraclePubKeyKey(oracleId uint64, pubKey []byte) []byte {
	key := append(PendingOraclePubKeyKeyPrefix, sdk.Uint64ToBigEndian(oracleId)...)
	key = append(key, pubKey...)

	return key
}

func AgencyKey(id uint64) []byte {
	return append(AgencyKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func PendingAgencyPubKeyKey(agencyId uint64, pubKey []byte) []byte {
	key := append(PendingAgencyPubKeyKeyPrefix, sdk.Uint64ToBigEndian(agencyId)...)
	key = append(key, pubKey...)

	return key
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

func EventByPriceKey(price sdkmath.Int) []byte {
	return append(EventByPriceKeyPrefix, price.BigInt().Bytes()...)
}

func CurrentEventPriceKey(pair string) []byte {
	return append(CurrentEventPriceKeyPrefix, []byte(pair)...)
}

func AttestationKey(id uint64) []byte {
	return append(AttestationKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func AttestationByEventKey(eventId uint64) []byte {
	return append(AttestationByEventKeyPrefix, sdk.Uint64ToBigEndian(eventId)...)
}

func PriceKey(pair string) []byte {
	return append(PriceKeyPrefix, []byte(pair)...)
}
