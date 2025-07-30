package types

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "oracle"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName

	BTCUSD      = "BTCUSD"
	NULL_SYMBOL = "_"

	flagOracleEnable         = "oracle.enable"
	flagOracleBitcoinRpc     = "oracle.bitcoin_rpc"
	flagOracleBitcoinRpcUser = "oracle.bitcoin_rpc_user"
	flagOracleBitcoinRpcPass = "oracle.bitcoin_rpc_password"
	flagOracleBitcoinRpcPost = "oracle.http_post_mode"
	flagOracleBitcoinRpcSSL  = "oracle.disable_tls"
)

var (
	SupportedPairs = []string{BTCUSD}

	ParamsStoreKey = []byte{0x01}

	PriceKeyPrefix            = []byte{0x10}
	BitcoinHeaderPrefix       = []byte{0x11}
	BitcoinHeaderHeightPrefix = []byte{0x12} // prefix for each key to a block hash, for a height
	BitcoinBestBlockHeaderKey = []byte{0x13} // key for the best block height

	PRICE_CACHE    = make(map[string]map[string]Price) // symbol, exchange, price[]
	PriceMu        sync.RWMutex
	StartProviders = false
)

func PriceKey(symbol string) []byte {
	return append(PriceKeyPrefix, []byte(symbol)...)
}

func BitcoinHeaderKey(hash string) []byte {
	return append(BitcoinHeaderPrefix, []byte(hash)...)
}

func BitcoinBlockHeaderHeightKey(height int32) []byte {
	return append(BitcoinHeaderHeightPrefix, sdk.Uint64ToBigEndian(uint64(height))...)
}
