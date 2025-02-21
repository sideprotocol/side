package mempool

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	"lending-tests/btcutils/client/btcapi"
)

type UTXO struct {
	Txid   string `json:"txid"`
	Vout   int    `json:"vout"`
	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int    `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int64  `json:"block_time"`
	} `json:"status"`
	Value int64 `json:"value"`
}

// UTXOs is a slice of UTXO
type UTXOs []UTXO

func (c *Client) ListUnspent(address btcutil.Address) ([]*btcapi.UnspentOutput, error) {
	statusCode, resp, err := c.BaseClient.Request(http.MethodGet, fmt.Sprintf("%s/address/%s/utxo", c.MempoolAPI, address.EncodeAddress()), c.BaseClient.GetBaseOptions())
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query utxos, status code: %d, response: %s", statusCode, string(resp))
	}

	var utxos UTXOs
	err = json.Unmarshal(resp, &utxos)
	if err != nil {
		return nil, err
	}

	unspentOutputs := make([]*btcapi.UnspentOutput, 0)
	for _, utxo := range utxos {
		txHash, err := chainhash.NewHashFromStr(utxo.Txid)
		if err != nil {
			return nil, err
		}

		pkScript, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, err
		}

		unspentOutputs = append(unspentOutputs, &btcapi.UnspentOutput{
			Outpoint: wire.NewOutPoint(txHash, uint32(utxo.Vout)),
			Output:   wire.NewTxOut(utxo.Value, pkScript),
		})
	}

	return unspentOutputs, nil
}
