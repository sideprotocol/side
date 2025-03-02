package binance

import (
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/gorilla/websocket"
	"github.com/sideprotocol/side/x/oracle/types"
)

type Subscription struct {
	Stream string           `json:"stream"`
	Data   SubscriptionData `json:"data"`
}

type SubscriptionData struct {
	Event     string `json:"e"`
	EventTime uint64 `json:"E"`
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
)

func symbol(source string) string {
	if target, ok := SymbolMap[source]; ok {
		return target
	} else {
		return source
	}
}

func Subscribe(svrCtx *server.Context) error {
	url := "wss://stream.binance.com:443/stream?streams=btcusdt@miniTicker"
	c, re, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		svrCtx.Logger.Error("price provider connection", "url", url, "status", re.Status, "body", re.Body)
	}
	defer c.Close()

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

		// adaptor(steam)
		// svrCtx.Logger.Info("Websocket Received", "message", subscription, "symbol", subscription.Data.Symbol, "price", subscription.Data.Close)

		price := types.Price{
			Symbol: symbol(subscription.Data.Symbol),
			Price:  subscription.Data.Close,
			Time:   subscription.Data.EventTime,
		}
		types.CachePrice(ProviderName, price)
	}
}
