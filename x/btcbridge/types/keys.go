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

	BtcBlockHeaderHashPrefix   = []byte{0x10} // prefix for each key to a block header, for a hash
	BtcBlockHeaderHeightPrefix = []byte{0x11} // prefix for each key to a block hash, for a height
	BtcBestBlockHeaderKey      = []byte{0x12} // key for the best block height
	BtcFeeRateKey              = []byte{0x13} // key for the bitcoin network fee rate

	BtcWithdrawRequestSequenceKey       = []byte{0x20} // key for the withdrawal request sequence
	BtcWithdrawRequestKeyPrefix         = []byte{0x21} // prefix for each key to a withdrawal request
	BtcWithdrawRequestByTxHashKeyPrefix = []byte{0x22} // prefix for each key to a withdrawal request by tx hash
	BtcWithdrawRequestQueueKeyPrefix    = []byte{0x23} // prefix for each key to a pending btc withdrawal request
	BtcSigningRequestSequenceKey        = []byte{0x24} // key for the signing request sequence
	BtcSigningRequestPrefix             = []byte{0x25} // prefix for each key to a signing request
	BtcSigningRequestByTxHashPrefix     = []byte{0x26} // prefix for each key to a signing request from tx hash
	BtcMintedTxHashKeyPrefix            = []byte{0x27} // prefix for each key to a minted tx hash

	BtcUtxoKeyPrefix              = []byte{0x30} // prefix for each key to a utxo
	BtcOwnerUtxoKeyPrefix         = []byte{0x31} // prefix for each key to an owned utxo
	BtcOwnerUtxoByAmountKeyPrefix = []byte{0x32} // prefix for each key to an owned utxo by amount
	BtcOwnerRunesUtxoKeyPrefix    = []byte{0x33} // prefix for each key to an owned runes utxo

	DKGRequestIDKey               = []byte{0x40} // key for the DKG request id
	DKGRequestKeyPrefix           = []byte{0x41} // prefix for each key to a DKG request
	DKGCompletionRequestKeyPrefix = []byte{0x42} // prefix for each key to a DKG completion request
	VaultVersionKey               = []byte{0x43} // key for vault version increased by 1 once updated
)

func BtcBlockHeaderHashKey(hash string) []byte {
	return append(BtcBlockHeaderHashPrefix, []byte(hash)...)
}

func BtcBlockHeaderHeightKey(height uint64) []byte {
	return append(BtcBlockHeaderHeightPrefix, sdk.Uint64ToBigEndian(height)...)
}

func BtcWithdrawRequestKey(sequence uint64) []byte {
	return append(BtcWithdrawRequestKeyPrefix, sdk.Uint64ToBigEndian(sequence)...)
}

func BtcWithdrawRequestByTxHashKey(txid string, sequence uint64) []byte {
	return append(append(BtcWithdrawRequestByTxHashKeyPrefix, []byte(txid)...), sdk.Uint64ToBigEndian(sequence)...)
}

func BtcWithdrawRequestQueueKey(sequence uint64) []byte {
	return append(BtcWithdrawRequestQueueKeyPrefix, sdk.Uint64ToBigEndian(sequence)...)
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

func BtcUtxoKey(hash string, vout uint64) []byte {
	return append(append(BtcUtxoKeyPrefix, []byte(hash)...), sdk.Uint64ToBigEndian(vout)...)
}

func BtcOwnerUtxoKey(owner string, hash string, vout uint64) []byte {
	key := append(append(BtcOwnerUtxoKeyPrefix, []byte(owner)...), []byte(hash)...)
	key = append(key, sdk.Uint64ToBigEndian(vout)...)

	return key
}

func BtcOwnerUtxoByAmountKey(owner string, amount uint64, hash string, vout uint64) []byte {
	key := append(append(BtcOwnerUtxoByAmountKeyPrefix, []byte(owner)...), sdk.Uint64ToBigEndian(amount)...)
	key = append(key, []byte(hash)...)
	key = append(key, sdk.Uint64ToBigEndian(vout)...)

	return key
}

func BtcOwnerRunesUtxoKey(owner string, id string, amount string, hash string, vout uint64) []byte {
	key := append(append(BtcOwnerRunesUtxoKeyPrefix, []byte(owner)...), MarshalRuneIdFromString(id)...)
	key = append(key, MarshalRuneAmountFromString(amount)...)
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
