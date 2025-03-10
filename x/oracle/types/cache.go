package types

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
	PriceMu.Lock()
	defer PriceMu.Unlock()
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
		if len(list) > 100 {
			list = list[100:]
		}
		list = append(list, p)
		target[ex] = list
	} else {
		target[ex] = []Price{p}
	}
}
