package types

import (
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/wire"

	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
)

const (
	// default tx version
	TxVersion = 2

	// default maximum allowed transaction weight
	MaxTransactionWeight = 400000
)

// BuildPsbt builds a psbt from the given params
func BuildPsbt(utxos []*btcbridgetypes.UTXO, recipientPkScript []byte, feeRate int64, witnessSize int) (*psbt.Packet, error) {
	txOut := wire.NewTxOut(0, recipientPkScript)

	unsignedTx, err := BuildUnsignedTransaction(utxos, txOut, feeRate, witnessSize)
	if err != nil {
		return nil, err
	}

	p, err := psbt.NewFromUnsignedTx(unsignedTx)
	if err != nil {
		return nil, err
	}

	for i, utxo := range utxos {
		p.Inputs[i].WitnessUtxo = wire.NewTxOut(int64(utxo.Amount), utxo.PubKeyScript)
	}

	return p, nil
}

// BuildUnsignedTransaction builds an unsigned tx from the given params
func BuildUnsignedTransaction(utxos []*btcbridgetypes.UTXO, txOut *wire.TxOut, feeRate int64, witnessSize int) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(TxVersion)

	inAmount := int64(0)
	outAmount := txOut.Value

	for _, utxo := range utxos {
		AddUTXOToTx(tx, utxo)
		inAmount += int64(utxo.Amount)
	}

	tx.AddTxOut(txOut)

	fee := GetTxVirtualSize(tx, witnessSize) * feeRate

	change := inAmount - outAmount - fee
	if change <= 0 {
		return nil, ErrInsufficientUTXOs
	}

	txOut.Value += change
	if btcbridgetypes.IsDustOut(txOut) {
		return nil, ErrDustOutput
	}

	if err := CheckTransactionWeight(tx, witnessSize); err != nil {
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

// GetTxVirtualSize gets the virtual size of the given tx.
func GetTxVirtualSize(tx *wire.MsgTx, witnessSize int) int64 {
	newTx := PopulateTxWithDummyWitness(tx, witnessSize)

	return mempool.GetTxVirtualSize(btcutil.NewTx(newTx))
}

// CheckTransactionWeight checks if the weight of the given tx exceeds the allowed maximum weight
func CheckTransactionWeight(tx *wire.MsgTx, witnessSize int) error {
	newTx := PopulateTxWithDummyWitness(tx, witnessSize)

	weight := blockchain.GetTransactionWeight(btcutil.NewTx(newTx))
	if weight > MaxTransactionWeight {
		return ErrMaxTransactionWeightExceeded
	}

	return nil
}

// PopulateTxWithDummyWitness populates the given tx with the dummy witness
func PopulateTxWithDummyWitness(tx *wire.MsgTx, witnessSize int) *wire.MsgTx {
	newTx := tx.Copy()

	for _, txIn := range newTx.TxIn {
		if len(txIn.Witness) == 0 {
			dummyWitness := make([]byte, witnessSize)
			txIn.Witness = wire.TxWitness{dummyWitness}
		}
	}

	return newTx
}
