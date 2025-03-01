package coinbase

import (
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/gorilla/websocket"
	"github.com/sideprotocol/side/x/oracle/types"
)

// ▼ {"type":"ticker","sequence":1566204232,"product_id":"BTC-USD","price":"1.11","open_24h":"1","volume_24h":"318996.26881102","low_24h":"0.38","high_24h":"1.48","volume_30d":"3743179.26781576","best_bid":"1.09","best_bid_size":"11.99966940","best_ask":"1.11","best_ask_size":"16.34610601","side":"buy","time":"2025-03-01T13:25:17.042052Z","trade_id":130278532,"last_size":"1.98198198"}

var (
	ProviderName = "coinbase"
	SymbolMap    = map[string]string{
		"BTC-USD": types.BTCUSD,
	}
)

func symbol(source string) string {
	if target, ok := SymbolMap[source]; ok {
		return target
	} else {
		return source
	}
}

type Subscription struct {
	Type   string `json:"type"`
	Symbol string `json:"product_id,omitempty"`
	Price  string `json:"price,omitempty"`
	Time   string `json:"time,omitempty"`
}

// {"type":"subscribe","product_ids":["BTC-USD"],"channels":[{"name":"ticker","product_ids":["BTC-USD"]}]}
func subscribe(conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"subscribe\",\"product_ids\":[\"BTC-USD\"],\"channels\":[{\"name\":\"ticker\",\"product_ids\":[\"BTC-USD\"]}]}"))
}

func Subscribe(svrCtx *server.Context) {
	// url := "wss://ws-feed-public.sandbox.exchange.coinbase.com"
	url := "wss://ws-feed.exchange.coinbase.com"
	c, re, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		svrCtx.Logger.Error("price provider connection", "url", url, "status", re.Status, "body", re.Body)
	}
	defer c.Close()

	subscribe(c)

	for {
		subscription := &Subscription{}
		err := c.ReadJSON(subscription)
		if err != nil {
			svrCtx.Logger.Error("reconnect websocket", "url", url, "error", err)
			time.Sleep(5 * time.Second)
			c, _, err = websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				svrCtx.Logger.Error("price provider connection", "url", url, "status", re.Status, "body", re.Body)
			}
		}

		if subscription.Type == "ticker" {
			// svrCtx.Logger.Info("Websocket Received", "provider", ProviderName, "message", subscription, "symbol", subscription.Symbol, "price", subscription.Price)

			// sample time: 2025-03-01T03:42:43.951417Z
			if t, err := time.Parse(time.RFC3339Nano, subscription.Time); err == nil {
				price := types.Price{
					Symbol: symbol(subscription.Symbol),
					Price:  subscription.Price,
					Time:   uint64(t.UnixMilli()),
				}
				types.CachePrice(ProviderName, price)
			} else {
				svrCtx.Logger.Error("Parse time error")
			}

		}
		// adaptor(steam)

	}
}
