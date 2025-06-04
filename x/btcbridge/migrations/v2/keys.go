package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// v1 keys
var (
	BtcBlockHeaderPrefix       = []byte{0x10} // prefix for each key to a block header by hash
	BtcBlockHeaderHeightPrefix = []byte{0x11} // prefix for each key to a block hash by height
	BtcBestBlockHeaderKey      = []byte{0x12} // key for the best block height
)

func BtcBlockHeaderHashKey(hash string) []byte {
	return append(BtcBlockHeaderPrefix, []byte(hash)...)
}

func BtcBlockHeaderHeightKey(height uint64) []byte {
	return append(BtcBlockHeaderHeightPrefix, sdk.Uint64ToBigEndian(height)...)
}
