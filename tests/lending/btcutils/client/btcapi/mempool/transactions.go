package mempool

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"lending-tests/btcutils/types"
)

func (c *Client) GetRawTransaction(txHash *chainhash.Hash) (*wire.MsgTx, error) {
	_, resp, err := c.BaseClient.Request(http.MethodGet, fmt.Sprintf("%s/tx/%s/raw", c.MempoolAPI, txHash.String()), c.BaseClient.GetBaseOptions())
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(types.TxVersion)
	if err := tx.Deserialize(bytes.NewReader(resp)); err != nil {
		return nil, err
	}

	return tx, nil
}

func (c *Client) BroadcastTx(tx *wire.MsgTx) (*chainhash.Hash, error) {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, err
	}

	options := c.BaseClient.GetBaseOptions()
	options.Body = []byte(hex.EncodeToString(buf.Bytes()))

	statusCode, resp, err := c.BaseClient.Request(http.MethodPost, fmt.Sprintf("%s/tx", c.MempoolAPI), options)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to broadcast tx, status code: %d, response: %s", statusCode, string(resp))
	}

	txHash, err := chainhash.NewHashFromStr(string(resp))
	if err != nil {
		return nil, fmt.Errorf("failed to parse tx hash: %s, err: %v", string(resp), err)
	}

	return txHash, nil
}
