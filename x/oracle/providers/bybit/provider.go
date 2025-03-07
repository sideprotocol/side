package bybit

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/gorilla/websocket"
	"github.com/sideprotocol/side/x/oracle/types"
)

// ▼ {"topic":"tickers.BTCUSDT","ts":1740835319867,"type":"snapshot","cs":64027791513,"data":{"symbol":"BTCUSDT","lastPrice":"84547.99","highPrice24h":"86584.84","lowPrice24h":"80609.26","prevPrice24h":"80900","volume24h":"9416.834648","turnover24h":"792190557.2149933","price24hPcnt":"0.0451","usdIndexPrice":"84501.308024"}}

var (
	ProviderName = "bybit"
	SymbolMap    = map[string]string{
		"BTCUSDT": types.BTCUSD,
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
	Topic string           `json:"topic"`
	Time  int64            `json:"ts"`
	Data  SubscriptionData `json:"data"`
}

type SubscriptionData struct {
	Symbol string `json:"symbol"`
	Price  string `json:"lastPrice"`
}

// {"type":"subscribe","product_ids":["BTC-USD"],"channels":[{"name":"ticker","product_ids":["BTC-USD"]}]}
func subscribe(conn *websocket.Conn) {
	msg := `{
    "op": "subscribe",
    "args": [
        "tickers.BTCUSDT",
		"tickers.ATOMUSDT"
    ]
}`
	conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func Subscribe(svrCtx *server.Context) error {
	// url := "wss://stream-testnet.bybit.com/v5/public/spot"
	url := "wss://stream.bybit.com/v5/public/spot"
	c, re, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		svrCtx.Logger.Error("price provider connection", "url", url)
		return err
	}
	defer c.Close()

	subscribe(c)
	reconnect := false

	for {

		if reconnect {
			for {
				time.Sleep(5 * time.Second)
				if c, _, err = websocket.DefaultDialer.Dial(url, nil); err == nil {
					reconnect = false
					subscribe(c)
					svrCtx.Logger.Info("reconnected price provider", "url", url, "status", re.Status, "body", re.Body)
					break
				}
			}
		}

		if _, b, err := c.ReadMessage(); err == nil {
			text := string(b)
			if strings.Contains(text, "topic") {

				subscription := &Subscription{}
				if err = json.Unmarshal(b, subscription); err == nil {
					// svrCtx.Logger.Info("Websocket Received", "provider", ProviderName, "symbol", subscription.Data.Symbol, "price", subscription.Data.Price)

					price := types.Price{
						Symbol: symbol(subscription.Data.Symbol),
						Price:  subscription.Data.Price,
						Time:   subscription.Time,
					}
					types.CachePrice(ProviderName, price)

				}

			}

		} else {
			c.Close()
			svrCtx.Logger.Error("Read Error", "error", err, "provider", ProviderName)
			reconnect = true

		}

	}
}
