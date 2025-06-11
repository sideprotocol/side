package types

import (
	"context"
	time "time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/gorilla/websocket"
)

func sendMessage(conn *websocket.Conn, msg string) {
	if len(msg) > 0 {
		conn.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}

func close(c *websocket.Conn) {
	if c != nil {
		c.Close()
	}
}

func Subscribe(provider string, svrCtx *server.Context, ctx context.Context, url, msg string, priceHander func(msg []byte) []Price) error {

	reconnect := true
	var c *websocket.Conn
	var err error
	defer close(c)

	for {
		if reconnect {
			for {
				time.Sleep(5 * time.Second)
				if c, _, err = websocket.DefaultDialer.Dial(url, nil); err == nil {
					reconnect = false
					sendMessage(c, msg)
					svrCtx.Logger.With("module", ModuleName).Info("connected price provider", "url", url)
					break
				} else {
					svrCtx.Logger.With("module", ModuleName).Error("re-connecting...", "error", err, "provider", provider)
				}
			}
		}

		if _, b, err := c.ReadMessage(); err == nil {
			prices := priceHander(b)
			for _, p := range prices {
				CachePrice(provider, p)
			}
		} else {
			svrCtx.Logger.With("module", ModuleName).Error("provider disconnected", "error", err, "provider", provider)
			c.Close()
			reconnect = true
		}
	}

}
