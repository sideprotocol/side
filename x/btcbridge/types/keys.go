package types

import (
	"math/big"
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

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var (
	ParamsStoreKey = []byte{0x1}
	SequenceKey    = []byte{0x2}

	// Host chain keys prefix the HostChain structs
	BtcBlockHeaderHashPrefix         = []byte{0x11} // prefix for each key to a block header, for a hash
	BtcBlockHeaderHeightPrefix       = []byte{0x12} // prefix for each key to a block hash, for a height
	BtcBestBlockHeaderKey            = []byte{0x13} // key for the best block height
	BtcWithdrawRequestPrefix         = []byte{0x14} // prefix for each key to a withdrawal request
	BtcWithdrawRequestByTxHashPrefix = []byte{0x15} // prefix for each key to a withdrawal request from tx hash
	BtcMintedTxHashKeyPrefix         = []byte{0x16} // prefix for each key to a minted tx hash

	DKGRequestIDKey               = []byte{0x20} // key for the DKG request id
	DKGRequestKeyPrefix           = []byte{0x21} // prefix for each key to a DKG request
	DKGCompletionRequestKeyPrefix = []byte{0x22} // prefix for each key to a DKG completion request
	VaultVersionKey               = []byte{0x23} // key for vault version increased by 1 once updated
)

func Int64ToBytes(number uint64) []byte {
	big := new(big.Int)
	big.SetUint64(number)
	return big.Bytes()
}

func BtcBlockHeaderHashKey(hash string) []byte {
	return append(BtcBlockHeaderHashPrefix, []byte(hash)...)
}

func BtcBlockHeaderHeightKey(height uint64) []byte {
	return append(BtcBlockHeaderHeightPrefix, Int64ToBytes(height)...)
}

func BtcWithdrawRequestKey(sequence uint64) []byte {
	return append(BtcWithdrawRequestPrefix, Int64ToBytes(sequence)...)
}

func BtcWithdrawRequestByTxHashKey(txid string) []byte {
	return append(BtcWithdrawRequestByTxHashPrefix, []byte(txid)...)
}

func BtcMintedTxHashKey(hash string) []byte {
	return append(BtcMintedTxHashKeyPrefix, []byte(hash)...)
}

func DKGRequestKey(id uint64) []byte {
	return append(DKGRequestKeyPrefix, Int64ToBytes(id)...)
}

func DKGCompletionRequestKey(id uint64, validator string) []byte {
	return append(append(DKGCompletionRequestKeyPrefix, Int64ToBytes(id)...), []byte(validator)...)
}
