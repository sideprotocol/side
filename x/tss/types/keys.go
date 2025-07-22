package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "tss"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_tss"
)

var (
	ParamsKey              = []byte{0x01} // key for params
	DKGRequestIdKey        = []byte{0x02} // key for dkg request id
	SigningRequestIdKey    = []byte{0x03} // key for signing request id
	RefreshingRequestIdKey = []byte{0x04} // key for refreshing request id

	DKGRequestKeyPrefix             = []byte{0x10} // key prefix for the dkg request
	DKGRequestByStatusKeyPrefix     = []byte{0x11} // key prefix for the dkg request by status
	DKGCompletionKeyPrefix          = []byte{0x12} // key prefix for the dkg completion
	SigningRequestKeyPrefix         = []byte{0x13} // key prefix for the signing request
	SigningRequestByStatusKeyPrefix = []byte{0x14} // key prefix for the signing request by status
	RefreshingRequestKeyPrefix      = []byte{0x15} // key prefix for the refreshing request
	RefreshingCompletionKeyPrefix   = []byte{0x16} // key prefix for the refreshing completion
)

func DKGRequestKey(id uint64) []byte {
	return append(DKGRequestKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func DKGRequestByStatusKey(status DKGStatus, id uint64) []byte {
	key := append(DKGRequestByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...)

	return append(key, sdk.Uint64ToBigEndian(id)...)
}

func DKGCompletionKey(id uint64, consPubKey string) []byte {
	return append(append(DKGCompletionKeyPrefix, sdk.Uint64ToBigEndian(id)...), []byte(consPubKey)...)
}

func SigningRequestKey(id uint64) []byte {
	return append(SigningRequestKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func SigningRequestByStatusKey(status SigningStatus, id uint64) []byte {
	key := append(SigningRequestByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...)

	return append(key, sdk.Uint64ToBigEndian(id)...)
}

func RefreshingRequestKey(id uint64) []byte {
	return append(RefreshingRequestKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func RefreshingCompletionKey(id uint64, consPubKey string) []byte {
	return append(append(RefreshingCompletionKeyPrefix, sdk.Uint64ToBigEndian(id)...), []byte(consPubKey)...)
}
