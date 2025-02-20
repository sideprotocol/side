package btcapi

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type UnspentOutput struct {
	Outpoint *wire.OutPoint
	Output   *wire.TxOut
}

type BTCAPIClient interface {
	GetRawTransaction(txHash *chainhash.Hash) (*wire.MsgTx, error)
	BroadcastTx(tx *wire.MsgTx) (*chainhash.Hash, error)
	ListUnspent(address btcutil.Address) ([]*UnspentOutput, error)
}
