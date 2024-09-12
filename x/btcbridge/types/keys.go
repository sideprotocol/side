package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "btcbridge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_btcbridge"
)

var (
	ParamsStoreKey = []byte{0x1}
	SequenceKey    = []byte{0x2}

	// Host chain keys prefix the HostChain structs
	BtcBlockHeaderHashPrefix        = []byte{0x11} // prefix for each key to a block header, for a hash
	BtcBlockHeaderHeightPrefix      = []byte{0x12} // prefix for each key to a block hash, for a height
	BtcBestBlockHeaderKey           = []byte{0x13} // key for the best block height
	BtcSigningRequestPrefix         = []byte{0x14} // prefix for each key to a signing request
	BtcSigningRequestByTxHashPrefix = []byte{0x15} // prefix for each key to a signing request from tx hash
	BtcMintedTxHashKeyPrefix        = []byte{0x16} // prefix for each key to a minted tx hash
	BtcLockedAssetKeyPrefix         = []byte{0x17} // prefix for each key to the locked asset

	BtcUtxoKeyPrefix           = []byte{0x20} // prefix for each key to a utxo
	BtcOwnerUtxoKeyPrefix      = []byte{0x21} // prefix for each key to an owned utxo
	BtcOwnerRunesUtxoKeyPrefix = []byte{0x22} // prefix for each key to an owned runes utxo

	DKGRequestIDKey               = []byte{0x30} // key for the DKG request id
	DKGRequestKeyPrefix           = []byte{0x31} // prefix for each key to a DKG request
	DKGCompletionRequestKeyPrefix = []byte{0x32} // prefix for each key to a DKG completion request
	VaultVersionKey               = []byte{0x33} // key for vault version increased by 1 once updated
)

func BtcBlockHeaderHashKey(hash string) []byte {
	return append(BtcBlockHeaderHashPrefix, []byte(hash)...)
}

func BtcBlockHeaderHeightKey(height uint64) []byte {
	return append(BtcBlockHeaderHeightPrefix, sdk.Uint64ToBigEndian(height)...)
}

func BtcSigningRequestKey(sequence uint64) []byte {
	return append(BtcSigningRequestPrefix, sdk.Uint64ToBigEndian(sequence)...)
}

func BtcSigningRequestByTxHashKey(txid string) []byte {
	return append(BtcSigningRequestByTxHashPrefix, []byte(txid)...)
}

func BtcMintedTxHashKey(hash string) []byte {
	return append(BtcMintedTxHashKeyPrefix, []byte(hash)...)
}

func BtcLockedAssetKey(txHash string, index uint8) []byte {
	return append(append(BtcLockedAssetKeyPrefix, []byte(txHash)...), byte(index))
}

func BtcUtxoKey(hash string, vout uint64) []byte {
	return append(append(BtcUtxoKeyPrefix, []byte(hash)...), sdk.Uint64ToBigEndian(vout)...)
}

func BtcOwnerUtxoKey(owner string, hash string, vout uint64) []byte {
	key := append(append(BtcOwnerUtxoKeyPrefix, []byte(owner)...), []byte(hash)...)
	key = append(key, sdk.Uint64ToBigEndian(vout)...)

	return key
}

func BtcOwnerRunesUtxoKey(owner string, id string, hash string, vout uint64) []byte {
	key := append(append(BtcOwnerRunesUtxoKeyPrefix, []byte(owner)...), []byte(id)...)
	key = append(key, []byte(hash)...)
	key = append(key, sdk.Uint64ToBigEndian(vout)...)

	return key
}

func DKGRequestKey(id uint64) []byte {
	return append(DKGRequestKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func DKGCompletionRequestKey(id uint64, consAddress string) []byte {
	return append(append(DKGCompletionRequestKeyPrefix, sdk.Uint64ToBigEndian(id)...), []byte(consAddress)...)
}
