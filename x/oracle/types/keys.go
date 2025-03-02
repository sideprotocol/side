package types

import (
	"sync"

	"cosmossdk.io/math"
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

	BTCUSD = "BTCUSD"

	flagOracleEnable = "oracle.enable"
)

var (
	Percent        = math.NewInt(100)
	Permille       = math.NewInt(1000)
	ParamsStoreKey = []byte{0x1}

	PriceKey = []byte{0x07}

	PRICE_CACHE = make(map[string]map[string]Price)
	mu          sync.Mutex
)

type Price struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
	Time   uint64 `json:"time"`
}

func CachePrice(exchange string, price Price) {
	mu.Lock()
	defer mu.Unlock()
	if v, ok := PRICE_CACHE[price.Symbol]; ok {
		v[exchange] = price
	} else {
		v = make(map[string]Price)
		v[exchange] = price
		PRICE_CACHE[price.Symbol] = v
	}
}
