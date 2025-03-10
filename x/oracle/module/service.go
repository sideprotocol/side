package oracle

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/sideprotocol/side/x/oracle/providers/binance"
	"github.com/sideprotocol/side/x/oracle/providers/bitget"
	"github.com/sideprotocol/side/x/oracle/providers/bybit"
	"github.com/sideprotocol/side/x/oracle/providers/coinbase"
	"github.com/sideprotocol/side/x/oracle/providers/okex"
	"github.com/sideprotocol/side/x/oracle/types"
	"golang.org/x/sync/errgroup"
)

// Start Oracle Price Service
// Subscrible Prices from providers
func Start(svrCtx *server.Context, clientCtx client.Context, ctx context.Context, g *errgroup.Group) error {

	if types.StartProviders {

		svrCtx.Logger.Info("price service", "module", "oracle", "msg", "Start Oracle Price Subscriber")

		// go binance.Subscribe(svrCtx)
		// go okex.Subscribe(svrCtx)
		// go coinbase.Subscribe(svrCtx)
		// go bybit.Subscribe(svrCtx)
		// go bitget.Subscribe(svrCtx)
		g.Go(func() error { return binance.Subscribe(svrCtx) })
		g.Go(func() error { return okex.Subscribe(svrCtx) })
		g.Go(func() error { return coinbase.Subscribe(svrCtx) })
		g.Go(func() error { return bybit.Subscribe(svrCtx) })
		g.Go(func() error { return bitget.Subscribe(svrCtx) })
	} else {
		svrCtx.Logger.Warn("Price service is disabled. It is required if your node is a validator. ")
	}

	return nil

}
