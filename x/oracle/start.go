package oracle

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/sideprotocol/side/x/oracle/providers/binance"
	"github.com/sideprotocol/side/x/oracle/providers/bybit"
	"github.com/sideprotocol/side/x/oracle/providers/coinbase"
	"github.com/sideprotocol/side/x/oracle/providers/okex"
	"golang.org/x/sync/errgroup"
)

type Starter struct {
}

// Start Oracle Price Service
// Subscrible Prices from providers
func Start(svrCtx *server.Context, clientCtx client.Context, ctx context.Context, g *errgroup.Group) error {

	svrCtx.Logger.Info("service start", "module", "oracle", "msg", "Start Oracle Price Subscriber")

	// g.Go(func() error { return binance.Subscribe(svrCtx, ctx) })
	// g.Go(func() error { return okex.Subscribe(svrCtx) })
	// g.Go(func() error { return coinbase.Subscribe(svrCtx) })
	// g.Go(func() error { return bybit.Subscribe(svrCtx) })
	// g.Go(func() error { return bitget.Subscribe(svrCtx) })
	go binance.Subscribe(svrCtx)
	go okex.Subscribe(svrCtx)
	go coinbase.Subscribe(svrCtx)
	go bybit.Subscribe(svrCtx)

	return nil

}
