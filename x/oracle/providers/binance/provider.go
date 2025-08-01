package binance

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/server"

	"github.com/sideprotocol/side/x/oracle/types"
)

type Subscription struct {
	Stream string           `json:"stream"`
	Data   SubscriptionData `json:"data"`
}

type SubscriptionData struct {
	Event     string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Close     string `json:"c"`
	// o string
	// h string
	// l string
	// v string
	// q string
}

var (
	ProviderName = "binance"
	SymbolMap    = map[string]string{
		"BTCUSDT": types.BTCUSD,
	}
	URL          = "wss://stream.binance.com:443/stream?streams=btcusdt@miniTicker/ethbtc@miniTicker"
	SubscribeMsg = ""
)

func symbol(source string) string {
	if target, ok := SymbolMap[source]; ok {
		return target
	} else {
		return source
	}
}

func Subscribe(svrCtx *server.Context, ctx context.Context) error {
	return types.Subscribe(ProviderName, svrCtx, ctx, URL, SubscribeMsg, func(msg []byte) []types.Price {
		subscription := &Subscription{}
		prices := []types.Price{}
		if err := json.Unmarshal(msg, &subscription); err == nil {
			price := types.Price{
				Symbol: symbol(subscription.Data.Symbol),
				Price:  subscription.Data.Close,
				Time:   subscription.Data.EventTime,
			}
			prices = append(prices, price)
		}
		return prices
	})
}

// func Subscribe(svrCtx *server.Context) error {
// 	url := "wss://stream.binance.com:443/stream?streams=btcusdt@miniTicker/atomusdt@miniTicker"
// 	c, re, err := websocket.DefaultDialer.Dial(url, nil)
// 	if err != nil {
// 		svrCtx.Logger.Error("price provider connection", "url", url)
// 	}
// 	defer c.Close()

// 	reconnect := false
// 	for {
// 		if reconnect {
// 			for {
// 				time.Sleep(5 * time.Second)
// 				if c, _, err = websocket.DefaultDialer.Dial(url, nil); err == nil {
// 					reconnect = false
// 					svrCtx.Logger.Info("reconnected price provider", "url", url, "status", re.Status, "body", re.Body)
// 					break
// 				}
// 			}
// 		} else {
// 			subscription := &Subscription{}
// 			if err = c.ReadJSON(subscription); err == nil {
// 				price := types.Price{
// 					Symbol: symbol(subscription.Data.Symbol),
// 					Price:  subscription.Data.Close,
// 					Time:   subscription.Data.EventTime,
// 				}

// 				types.CachePrice(ProviderName, price)
// 			} else {
// 				svrCtx.Logger.Error("Price Read Error", "error", err, "provider", ProviderName)
// 				c.Close()
// 				reconnect = true
// 			}
// 		}
// 	}
// }
