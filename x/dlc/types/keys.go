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
	ParamsKey                   = []byte{0x01} // key for params
	OracleIdKey                 = []byte{0x02} // key for oracle id
	DCMIdKey                    = []byte{0x03} // key for DCM id
	EventIdKey                  = []byte{0x04} // key for event id
	PendingLendingEventCountKey = []byte{0x05} // key for pending lending event count
	AttestationIdKey            = []byte{0x06} // key for attestation id

	OracleKeyPrefix              = []byte{0x10} // prefix for each key to an oracle
	OracleByPubKeyKeyPrefix      = []byte{0x11} // prefix for each key to an oracle by public key
	DCMKeyPrefix                 = []byte{0x12} // prefix for each key to a DCM
	DCMByPubKeyKeyPrefix         = []byte{0x13} // prefix for each key to a DCM by public key
	NonceIndexKeyPrefix          = []byte{0x14} // key prefix for the nonce index
	NonceKeyPrefix               = []byte{0x15} // prefix for each key to a nonce
	NonceByValueKeyPrefix        = []byte{0x16} // key prefix for the nonce value
	EventKeyPrefix               = []byte{0x17} // prefix for each key to an event
	EventByStatusKeyPrefix       = []byte{0x18} // prefix for each key to an event by status
	PendingLendingEventKeyPrefix = []byte{0x19} // key prefix for the pending lending event
	AttestationKeyPrefix         = []byte{0x20} // prefix for each key to an attestation
	AttestationByEventKeyPrefix  = []byte{0x21} // prefix for each key to an attestation by event

	OracleParticipantLivenessKeyPrefix = []byte{0x30} // key prefix for oracle participant liveness
)

func OracleKey(id uint64) []byte {
	return append(OracleKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func OracleByPubKeyKey(pubKey []byte) []byte {
	return append(OracleByPubKeyKeyPrefix, pubKey...)
}

func DCMKey(id uint64) []byte {
	return append(DCMKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func DCMByPubKeyKey(pubKey []byte) []byte {
	return append(DCMByPubKeyKeyPrefix, pubKey...)
}

func NonceIndexKey(oracleId uint64) []byte {
	return append(NonceIndexKeyPrefix, sdk.Uint64ToBigEndian(oracleId)...)
}

func NonceKey(oracleId uint64, index uint64) []byte {
	return append(append(NonceKeyPrefix, sdk.Uint64ToBigEndian(oracleId)...), sdk.Uint64ToBigEndian(index)...)
}

func NonceByValueKey(nonce []byte) []byte {
	return append(NonceByValueKeyPrefix, nonce...)
}

func EventKey(id uint64) []byte {
	return append(EventKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func EventByStatusKey(triggered bool, id uint64) []byte {
	key := append(EventByStatusKeyPrefix, EventStatusToByte(triggered))

	return append(key, sdk.Uint64ToBigEndian(id)...)
}

func PendingLendingEventKey(id uint64) []byte {
	return append(PendingLendingEventKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func AttestationKey(id uint64) []byte {
	return append(AttestationKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func AttestationByEventKey(eventId uint64) []byte {
	return append(AttestationByEventKeyPrefix, sdk.Uint64ToBigEndian(eventId)...)
}

func OracleParticipantLivenessKey(consPubKey string) []byte {
	return append(OracleParticipantLivenessKeyPrefix, []byte(consPubKey)...)
}
