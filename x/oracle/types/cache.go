package types

import (
	"cosmossdk.io/math"
)

type Price struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
	Time   int64  `json:"time"`
}

func CachePrice(exchange string, price Price) {
	if len(price.Symbol) == 0 || len(price.Price) == 0 {
		return
	}

	PriceMu.Lock()
	defer PriceMu.Unlock()

	if v, ok := PRICE_CACHE[price.Symbol]; ok {
		v[exchange] = price
		PRICE_CACHE[price.Symbol] = v
	} else {
		v = make(map[string]Price)
		v[exchange] = price
		PRICE_CACHE[price.Symbol] = v
	}

	// a null price is used to identify whether price provider is working
	nullPrice := nullPrice(price.Time)
	if v, ok := PRICE_CACHE[nullPrice.Symbol]; ok {
		v[exchange] = nullPrice
		PRICE_CACHE[nullPrice.Symbol] = v
	} else {
		v = make(map[string]Price)
		v[exchange] = nullPrice
		PRICE_CACHE[nullPrice.Symbol] = v
	}
}

func nullPrice(pTime int64) Price {
	return Price{
		Symbol: NULL_SYMBOL,
		Price:  "0",
		Time:   pTime,
	}
}

func GetPrices(lastBlockTime int64) map[string][]math.LegacyDec {
	PriceMu.RLock()
	defer PriceMu.RUnlock()

	// calculate the weighted average
	symbolPrices := make(map[string][]math.LegacyDec)
	for symbol, pairs := range PRICE_CACHE {
		for _, price := range pairs {
			if price.Time > lastBlockTime {
				p, err := math.LegacyNewDecFromStr(price.Price)
				if err == nil {
					symbolPrices[symbol] = append(symbolPrices[symbol], p)
				}
			}
		}
	}
	return symbolPrices
}

// func setMapValue(target map[string][]Price, ex string, p Price) {
// 	if list, ok := target[ex]; ok {
// 		if len(list) > 100 {
// 			list = list[100:]
// 		}
// 		list = append(list, p)
// 		target[ex] = list
// 	} else {
// 		target[ex] = []Price{p}
// 	}
// }
