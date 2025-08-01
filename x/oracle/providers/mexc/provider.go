package mexc

import (
	"time"

	"github.com/gorilla/websocket"

	"github.com/cosmos/cosmos-sdk/server"

	"github.com/sideprotocol/side/x/oracle/types"
)

var (
	ProviderName = "mexc"
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
	Stream string           `json:"stream"`
	Data   SubscriptionData `json:"data"`
}

type SubscriptionData struct {
	Event     string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Close     string `json:"c"`
}

// { "method":"SUBSCRIPTION", "params":["spot@public.miniTicker.v3.api@BTCUSDT"] }
type Message struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
}

func subscribe(conn *websocket.Conn) {
	msg := Message{Method: "SUBSCRIPTION", Params: []string{"spot@public.miniTicker.v3.api@BTCUSDT"}}
	conn.WriteJSON(msg)
}

func Subscribe(svrCtx *server.Context) error {
	url := "ws://wbs-api.mexc.com/ws"
	c, re, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		svrCtx.Logger.Error("price provider connection", "url", url, "status", re.Status, "body", re.Body)
		return nil
	}
	defer c.Close()

	subscribe(c)
	reconnect := false
	for {

		if reconnect {
			svrCtx.Logger.Error("reconnect websocket", "url", url, "error", err)
			time.Sleep(5 * time.Second)
			c, _, err = websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				svrCtx.Logger.Error("price provider connection", "url", url, "status", re.Status, "body", re.Body)
			} else {
				reconnect = false
			}
		}

		subscription := &Subscription{}
		if err = c.ReadJSON(subscription); err == nil {
			svrCtx.Logger.Info("Websocket Received", "message", subscription, "symbol", subscription.Data.Symbol, "price", subscription.Data.Close)

			price := types.Price{
				Symbol: symbol(subscription.Data.Symbol),
				Price:  subscription.Data.Close,
				Time:   subscription.Data.EventTime,
			}
			types.CachePrice(ProviderName, price)
		} else {
			c.Close()
			svrCtx.Logger.Error("Read Error", "error", err, "provider", ProviderName)
			reconnect = true

		}

	}
}
