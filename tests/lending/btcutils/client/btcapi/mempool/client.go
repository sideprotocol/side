package mempool

import (
	"log"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"

	"lending-tests/btcutils/client/base"
	"lending-tests/btcutils/client/btcapi"
)

var _ btcapi.BTCAPIClient = (*Client)(nil)

// Client defines the mempool client
type Client struct {
	BaseClient *base.Client

	MempoolAPI string
}

// NewClient creates a mempool client instance
func NewClient(netParams *chaincfg.Params, baseClient *base.Client) *Client {
	mempoolAPI := ""

	if netParams.Net == wire.MainNet {
		mempoolAPI = "https://mempool.space/api"
	} else if netParams.Net == wire.TestNet3 {
		mempoolAPI = "https://mempool.space/testnet/api"
	} else if netParams.Net == chaincfg.SigNetParams.Net {
		mempoolAPI = "https://mempool.space/signet/api"
	} else {
		log.Fatal("mempool don't support other netParams")
	}

	return &Client{
		BaseClient: baseClient,
		MempoolAPI: mempoolAPI,
	}
}
