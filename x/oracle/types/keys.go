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
