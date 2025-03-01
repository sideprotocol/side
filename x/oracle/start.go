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
	"golang.org/x/sync/errgroup"
)

// var PRICE_CACHE = make(map[string]Price)

// Start Oracle Price Service
// Subscrible Prices from providers
func Start(svrCtx *server.Context, clientCtx client.Context, ctx context.Context, g *errgroup.Group) error {

	svrCtx.Logger.Info("service start", "module", "oracle", "msg", "Start Oracle Price Subscriber")

	go binance.Subscribe(svrCtx)
	go okex.Subscribe(svrCtx)
	go coinbase.Subscribe(svrCtx)
	go bybit.Subscribe(svrCtx)
	go bitget.Subscribe(svrCtx)

	return nil

}
