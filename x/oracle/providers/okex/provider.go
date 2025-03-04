package okex

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/gorilla/websocket"
	"github.com/sideprotocol/side/x/oracle/types"
)

// ▼ {"arg":{"channel":"index-tickers","instId":"BTC-USDT"},"data":[{"instId":"BTC-USDT","idxPx":"84631.5","open24h":"80469.1","high24h":"86556.5","low24h":"80079.3","sodUtc0":"84343.8","sodUtc8":"84008.9","ts":"1740830316879"}]}

var (
	ProviderName = "okex"
	SymbolMap    = map[string]string{
		"BTC-USDT": types.BTCUSD,
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
	Data []SubscriptionData `json:"data"`
}

type SubscriptionData struct {
	Symbol string `json:"instId"`
	Price  string `json:"idxPx"`
	Time   string `json:"ts"`
}

func subscribe(conn *websocket.Conn) {
	msg := `{
  "op": "subscribe",
  "args": [
    {
      "channel": "index-tickers",
      "instId": "BTC-USDT"
    }
  ]
}`
	conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func Subscribe(svrCtx *server.Context) error {
	url := "wss://ws.okx.com:8443/ws/v5/public"
	// url := "wss://wspap.okx.com:8443/ws/v5/public"
	c, re, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		svrCtx.Logger.Error("price provider connection", "url", url, "status", re.Status, "body", re.Body)
		return err
	}
	defer c.Close()

	subscribe(c)

	for {

		if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
			for {
				time.Sleep(10 * time.Second)
				if c, _, err = websocket.DefaultDialer.Dial(url, nil); err == nil {
					subscribe(c)
					svrCtx.Logger.Info("reconnected price provider", "url", url, "status", re.Status, "body", re.Body)
					break
				}
			}
		}

		if _, b, err := c.ReadMessage(); err == nil {
			text := string(b)
			if strings.Contains(text, "data") {
				subscription := &Subscription{}
				if err = json.Unmarshal(b, subscription); err == nil {

					for _, data := range subscription.Data {
						// svrCtx.Logger.Info("Websocket Received", "provider", ProviderName, "symbol", data.Symbol, "price", data.Price)

						if t, err := strconv.ParseInt(data.Time, 10, 64); err == nil {
							price := types.Price{
								Symbol: symbol(data.Symbol),
								Price:  data.Price,
								Time:   t,
							}
							types.CachePrice(ProviderName, price)
						}
					}

				}

			}
		}

	}
}
