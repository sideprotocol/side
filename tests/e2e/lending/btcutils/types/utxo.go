package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/tidwall/gjson"
)

// UTXO defines the utxo struct
type UTXO struct {
	Hash     chainhash.Hash `json:"hash"`
	Index    uint32         `json:"index"`
	Value    int64          `json:"value"`
	PkScript []byte         `json:"pkScript"`
}

// NewUTXO creates a new UTXO instance
func NewUTXO(hash *chainhash.Hash, index uint32, value int64, pkScript []byte) *UTXO {
	return &UTXO{
		Hash:     *hash,
		Index:    index,
		Value:    value,
		PkScript: pkScript,
	}
}

// GetOutPoint returns the outpoint from the UTXO
func (u *UTXO) GetOutPoint() *wire.OutPoint {
	return wire.NewOutPoint(&u.Hash, u.Index)
}

// GetOutput returns the tx output from the UTXO
func (u *UTXO) GetOutput() *wire.TxOut {
	return wire.NewTxOut(u.Value, u.PkScript)
}

// IsZeroHash returns true if the hash of the UTXO is zero value, false otherwise
func (u *UTXO) IsZeroHash() bool {
	return u.Hash.IsEqual(new(chainhash.Hash))
}

// UnmarshalJSON unmarshals the given data to the UTXO struct
func (u *UTXO) UnmarshalJSON(data []byte) error {
	if !gjson.ValidBytes(data) {
		return errors.New("invalid JSON")
	}

	json := gjson.ParseBytes(data)

	hash, err := chainhash.NewHashFromStr(json.Get("hash").String())
	if err != nil {
		return fmt.Errorf("invalid hash: %s", json.Get("hash").String())
	}

	pkScript, err := hex.DecodeString(json.Get("pkScript").String())
	if err != nil {
		return fmt.Errorf("invalid pk script: %s", json.Get("pkScript").String())
	}

	u.Hash = *hash
	u.Index = uint32(json.Get("index").Uint())
	u.Value = json.Get("value").Int()
	u.PkScript = pkScript

	return nil
}

// UTXOs is a set of UTXOs
type UTXOs []*UTXO

// TotalValue returns the total value of the given utxo set
func (us UTXOs) TotalValue() int64 {
	totalValue := int64(0)

	for _, utxo := range us {
		totalValue += utxo.Value
	}

	return totalValue
}

// String implements fmt.Stringer
func (us UTXOs) String() string {
	bz, err := json.Marshal(us)
	if err != nil {
		return ""
	}

	return string(bz)
}
