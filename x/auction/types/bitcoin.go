package types

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	errorsmod "cosmossdk.io/errors"

	"github.com/sideprotocol/side/bitcoin"
	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
)

const (
	// default tx version
	TxVersion = 2

	// default sig hash type
	DefaultSigHashType = txscript.SigHashDefault
)

// BuildPaymentTransaction builds the payment tx for the given auction and bids
func BuildPaymentTransaction(auction *Auction, bids []*Bid, feeRate int64) (string, *chainhash.Hash, []string, error) {
	liquidationCet, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(auction.LiquidationCet)), true)
	if err != nil {
		return "", nil, nil, err
	}

	txOut := liquidationCet.UnsignedTx.TxOut[0]

	utxo := &btcbridgetypes.UTXO{
		Txid:         liquidationCet.UnsignedTx.TxHash().String(),
		Vout:         0,
		Amount:       uint64(txOut.Value),
		PubKeyScript: txOut.PkScript,
	}

	paymentTxPsbt, err := BuildBatchTransferPsbt([]*btcbridgetypes.UTXO{utxo}, bids, feeRate, auction.Borrower)
	if err != nil {
		return "", nil, nil, err
	}

	paymentTxPsbtB64, err := paymentTxPsbt.B64Encode()
	if err != nil {
		return "", nil, nil, err
	}

	sigHashes := make([]string, 0)

	for i := range paymentTxPsbt.Inputs {
		sigHash, err := CalcTaprootSigHash(paymentTxPsbt, i, DefaultSigHashType)
		if err != nil {
			return "", nil, nil, err
		}

		sigHashes = append(sigHashes, hex.EncodeToString(sigHash))
	}

	txHash := paymentTxPsbt.UnsignedTx.TxHash()

	return paymentTxPsbtB64, &txHash, sigHashes, nil
}

// BuildBatchTransferPsbt builds the psbt to perform batch transfer to bidders
func BuildBatchTransferPsbt(utxos []*btcbridgetypes.UTXO, bids []*Bid, feeRate int64, change string) (*psbt.Packet, error) {
	chainCfg := bitcoin.Network

	txOuts := make([]*wire.TxOut, len(bids))

	for i, bid := range bids {
		address, err := btcutil.DecodeAddress(bid.Bidder, chainCfg)
		if err != nil {
			return nil, err
		}

		pkScript, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, err
		}

		txOuts[i] = wire.NewTxOut(int64(bid.BiddedAmount.Amount.Uint64()), pkScript)
	}

	changeAddress, err := btcutil.DecodeAddress(change, chainCfg)
	if err != nil {
		return nil, err
	}

	changePkScript, err := txscript.PayToAddrScript(changeAddress)
	if err != nil {
		return nil, err
	}

	unsignedTx, err := BuildUnsignedTransaction(utxos, txOuts, feeRate, changePkScript)
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
func BuildUnsignedTransaction(utxos []*btcbridgetypes.UTXO, txOuts []*wire.TxOut, feeRate int64, changePkScript []byte) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(TxVersion)

	inAmount := int64(0)
	outAmount := int64(0)

	for _, utxo := range utxos {
		AddUTXOToTx(tx, utxo)
		inAmount += int64(utxo.Amount)
	}

	for _, out := range txOuts {
		tx.AddTxOut(out)
		outAmount += out.Value
	}

	tx.AddTxOut(wire.NewTxOut(0, changePkScript))

	fee := btcbridgetypes.GetTxVirtualSize(tx, utxos) * feeRate

	changeValue := inAmount - outAmount - fee
	if changeValue > 0 {
		tx.TxOut[len(tx.TxOut)-1].Value = changeValue
		if btcbridgetypes.IsDustOut(tx.TxOut[len(tx.TxOut)-1]) {
			tx.TxOut = tx.TxOut[0 : len(tx.TxOut)-1]
		}
	} else {
		tx.TxOut = tx.TxOut[0 : len(tx.TxOut)-1]

		if changeValue < 0 {
			feeWithoutChange := btcbridgetypes.GetTxVirtualSize(tx, utxos) * feeRate
			if inAmount-outAmount-feeWithoutChange < 0 {
				return nil, errorsmod.Wrap(ErrFailedToBuildTx, "insufficient utxos")
			}
		}
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

// CalcTaprootSigHash computes the sig hash of the given input
// Assume that the psbt is valid
func CalcTaprootSigHash(p *psbt.Packet, idx int, sigHashType txscript.SigHashType) ([]byte, error) {
	prevOutFetcher := txscript.NewMultiPrevOutFetcher(nil)
	for i, txIn := range p.UnsignedTx.TxIn {
		prevOutFetcher.AddPrevOut(txIn.PreviousOutPoint, p.Inputs[i].WitnessUtxo)
	}

	sigHash, err := txscript.CalcTaprootSignatureHash(txscript.NewTxSigHashes(p.UnsignedTx, prevOutFetcher), sigHashType, p.UnsignedTx, idx, prevOutFetcher)
	if err != nil {
		return nil, err
	}

	return sigHash, nil
}
