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
	go func() {
		reconnect := true
		var c *websocket.Conn
		var err error
		defer close(c)

		if reconnect {
			for {
				time.Sleep(5 * time.Second)
				if c, _, err = websocket.DefaultDialer.Dial(url, nil); err == nil {
					reconnect = false
					sendMessage(c, msg)
					svrCtx.Logger.Info("reconnected price provider", "url", url)
					break
				}
			}
		}

		if _, b, err := c.ReadMessage(); err == nil {
			prices := priceHander(b)
			for _, p := range prices {
				CachePrice(provider, p)
			}
		} else {
			svrCtx.Logger.Error("Read Error", "error", err, "provider", provider)
			c.Close()
			reconnect = true
		}
	}()

	<-ctx.Done()

	svrCtx.Logger.Info("service stop", "module", ModuleName, "provider", provider)
	return nil
}
