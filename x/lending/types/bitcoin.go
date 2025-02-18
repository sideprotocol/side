package types

import (
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
)

const (
	// default tx version
	TxVersion = 2

	// default sig hash type
	DefaultSigHashType = txscript.SigHashDefault
)

// BuildPsbt builds a psbt from the given params
func BuildPsbt(utxos []*btcbridgetypes.UTXO, recipientPkScript []byte, feeRate int64) (*psbt.Packet, error) {
	txOut := wire.NewTxOut(0, recipientPkScript)

	unsignedTx, err := BuildUnsignedTransaction(utxos, txOut, feeRate)
	if err != nil {
		return nil, err
	}

	p, err := psbt.NewFromUnsignedTx(unsignedTx)
	if err != nil {
		return nil, err
	}

	for i, utxo := range utxos {
		p.Inputs[i].SighashType = DefaultSigHashType
		p.Inputs[i].WitnessUtxo = wire.NewTxOut(int64(utxo.Amount), utxo.PubKeyScript)
	}

	return p, nil
}

// BuildUnsignedTransaction builds an unsigned tx from the given params
func BuildUnsignedTransaction(utxos []*btcbridgetypes.UTXO, txOut *wire.TxOut, feeRate int64) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(TxVersion)

	inAmount := int64(0)
	outAmount := txOut.Value

	for _, utxo := range utxos {
		AddUTXOToTx(tx, utxo)
		inAmount += int64(utxo.Amount)
	}

	tx.AddTxOut(txOut)

	fee := btcbridgetypes.GetTxVirtualSize(tx, utxos) * feeRate

	change := inAmount - outAmount - fee
	if change <= 0 {
		return nil, ErrFailedToBuildTx
	}

	txOut.Value += change
	if btcbridgetypes.IsDustOut(txOut) {
		return nil, ErrFailedToBuildTx
	}

	if err := btcbridgetypes.CheckTransactionWeight(tx, utxos); err != nil {
		return nil, err
	}

	return tx, nil
}

// AddUTXOToTx adds the given utxo to the specified tx
// Make sure the utxo is valid
func AddUTXOToTx(tx *wire.MsgTx, utxo *btcbridgetypes.UTXO) {
	txIn := new(wire.TxIn)

	hash, err := chainhash.NewHashFromStr(utxo.Txid)
	if err != nil {
		panic(err)
	}

	txIn.PreviousOutPoint = *wire.NewOutPoint(hash, uint32(utxo.Vout))

	tx.AddTxIn(txIn)
}
