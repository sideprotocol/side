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

	PRICE_CACHE = make(map[string]map[string][]Price) // symbol, exchange, price[]
	mu          sync.Mutex
)

type Price struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
	Time   int64  `json:"time"`
}

func CachePrice(exchange string, price Price) {
	mu.Lock()
	defer mu.Unlock()
	if v, ok := PRICE_CACHE[price.Symbol]; ok {
		// v[exchange] = price
		setMapValue(v, exchange, price)
		PRICE_CACHE[price.Symbol] = v
	} else {
		v = make(map[string][]Price)
		setMapValue(v, exchange, price)
		PRICE_CACHE[price.Symbol] = v

	}
}

func CleanPrices(expire int64) {
	mu.Lock()
	defer mu.Unlock()
	for symbol, v := range PRICE_CACHE {
		for ex, list := range v {
			newList := []Price{}
			for _, p := range list {
				if p.Time > expire {
					newList = append(newList, p)
				}
			}
			PRICE_CACHE[symbol][ex] = newList
		}
	}
}

func setMapValue(target map[string][]Price, ex string, p Price) {
	if list, ok := target[ex]; ok {
		if len(list) > 500 {
			list = list[500:]
		}
		list = append(list, p)
		target[ex] = list
	} else {
		target[ex] = []Price{p}
	}
}
