package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "auction"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_auction"
)

var (
	ParamsKey    = []byte{0x01} // key for params
	AuctionIdKey = []byte{0x02} // key for auction id
	BidIdKey     = []byte{0x03} // key for bid id

	AuctionKeyPrefix = []byte{0x10} // prefix for each key to an auction
	BidKeyPrefix     = []byte{0x11} // prefix for each key to a bid
)

func AuctionKey(id uint64) []byte {
	return append(AuctionKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

func BidKey(id uint64) []byte {
	return append(BidKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}
