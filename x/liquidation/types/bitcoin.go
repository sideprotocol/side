package types

import (
	"bytes"
	"encoding/base64"
	"errors"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	errorsmod "cosmossdk.io/errors"

	"github.com/sideprotocol/side/bitcoin"
	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
)

const (
	// default tx version
	TxVersion = 2

	// default sequence
	DefaultSequence = 0

	// default sig hash type
	DefaultSigHashType = txscript.SigHashDefault
)

// BuildSettlementTransaction builds the settlement tx for the given liquidation and records
// If the fee rate is 0 or too high, the reserved network fee will be used instead
func BuildSettlementTransaction(liquidation *Liquidation, records []*LiquidationRecord, protocolFeeCollector string, feeRate int64, reservedNetworkFee int64) (string, *chainhash.Hash, []string, int64, error) {
	liquidationCet, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidation.LiquidationCet)), true)
	if err != nil {
		return "", nil, nil, 0, err
	}

	txOut := liquidationCet.UnsignedTx.TxOut[0]

	utxo := &btcbridgetypes.UTXO{
		Txid:         liquidationCet.UnsignedTx.TxHash().String(),
		Vout:         0,
		Amount:       uint64(txOut.Value),
		PubKeyScript: txOut.PkScript,
	}

	settlementTxPsbt, changeAmount, err := BuildBatchTransferPsbt([]*btcbridgetypes.UTXO{utxo}, records, protocolFeeCollector, liquidation.ProtocolLiquidationFee.Amount.Int64(), feeRate, reservedNetworkFee, liquidation.Debtor)
	if err != nil {
		return "", nil, nil, 0, err
	}

	settlementTxPsbtB64, err := settlementTxPsbt.B64Encode()
	if err != nil {
		return "", nil, nil, 0, err
	}

	sigHashes := make([]string, 0)

	for i := range settlementTxPsbt.Inputs {
		sigHash, err := CalcTaprootSigHash(settlementTxPsbt, i, DefaultSigHashType)
		if err != nil {
			return "", nil, nil, 0, err
		}

		sigHashes = append(sigHashes, base64.StdEncoding.EncodeToString(sigHash))
	}

	txHash := settlementTxPsbt.UnsignedTx.TxHash()

	return settlementTxPsbtB64, &txHash, sigHashes, changeAmount, nil
}

// BuildBatchTransferPsbt builds the psbt to perform batch transfer to liquidators, protocol fee collector and debtor(if remaining)
func BuildBatchTransferPsbt(utxos []*btcbridgetypes.UTXO, records []*LiquidationRecord, protocolFeeCollector string, protocolFee int64, feeRate int64, reservedNetworkFee int64, change string) (*psbt.Packet, int64, error) {
	chainCfg := bitcoin.Network

	txOuts := make([]*wire.TxOut, 0)

	for _, record := range records {
		address, err := btcutil.DecodeAddress(record.Liquidator, chainCfg)
		if err != nil {
			return nil, 0, err
		}

		pkScript, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, 0, err
		}

		txOuts = append(txOuts, wire.NewTxOut(record.CollateralAmount.Add(record.BonusAmount).Amount.Int64(), pkScript))
	}

	protocolFeeCollectorAddr, err := btcutil.DecodeAddress(protocolFeeCollector, chainCfg)
	if err != nil {
		return nil, 0, err
	}

	protocolFeeCollectorPkScript, err := txscript.PayToAddrScript(protocolFeeCollectorAddr)
	if err != nil {
		return nil, 0, err
	}

	protocolFeeOut := wire.NewTxOut(protocolFee, protocolFeeCollectorPkScript)
	if !btcbridgetypes.IsDustOut(protocolFeeOut) {
		txOuts = append(txOuts, wire.NewTxOut(protocolFee, protocolFeeCollectorPkScript))
	}

	changeAddress, err := btcutil.DecodeAddress(change, chainCfg)
	if err != nil {
		return nil, 0, err
	}

	changePkScript, err := txscript.PayToAddrScript(changeAddress)
	if err != nil {
		return nil, 0, err
	}

	var unsignedTx *wire.MsgTx
	var changeAmount int64

	// try fee rate mode first
	if feeRate != 0 {
		unsignedTx, changeAmount, err = BuildUnsignedTransaction(utxos, txOuts, feeRate, changePkScript)
		if err != nil && !errors.Is(err, ErrInsufficientUTXOs) {
			return nil, 0, err
		}
	}

	// try fee mode then
	if unsignedTx == nil {
		unsignedTx, changeAmount, err = BuildUnsignedTransactionWithFee(utxos, txOuts, reservedNetworkFee, changePkScript)
		if err != nil {
			return nil, 0, err
		}
	}

	p, err := psbt.NewFromUnsignedTx(unsignedTx)
	if err != nil {
		return nil, 0, err
	}

	for i, utxo := range utxos {
		p.Inputs[i].SighashType = DefaultSigHashType
		p.Inputs[i].WitnessUtxo = wire.NewTxOut(int64(utxo.Amount), utxo.PubKeyScript)
	}

	return p, changeAmount, nil
}

// BuildUnsignedTransaction builds an unsigned tx from the given params
func BuildUnsignedTransaction(utxos []*btcbridgetypes.UTXO, txOuts []*wire.TxOut, feeRate int64, changePkScript []byte) (*wire.MsgTx, int64, error) {
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

	changeAmount := inAmount - outAmount - fee
	if changeAmount > 0 {
		tx.TxOut[len(tx.TxOut)-1].Value = changeAmount
		if btcbridgetypes.IsDustOut(tx.TxOut[len(tx.TxOut)-1]) {
			tx.TxOut = tx.TxOut[0 : len(tx.TxOut)-1]
			changeAmount = 0
		}
	} else {
		tx.TxOut = tx.TxOut[0 : len(tx.TxOut)-1]

		if changeAmount < 0 {
			feeWithoutChange := btcbridgetypes.GetTxVirtualSize(tx, utxos) * feeRate
			if inAmount-outAmount-feeWithoutChange < 0 {
				return nil, 0, ErrInsufficientUTXOs
			}
		}

		changeAmount = 0
	}

	if err := btcbridgetypes.CheckTransactionWeight(tx, utxos); err != nil {
		return nil, 0, err
	}

	return tx, changeAmount, nil
}

// BuildUnsignedTransactionWithFee is similar to BuildUnsignedTransaction, except that it uses the specified fee instead of the fee rate
func BuildUnsignedTransactionWithFee(utxos []*btcbridgetypes.UTXO, txOuts []*wire.TxOut, fee int64, changePkScript []byte) (*wire.MsgTx, int64, error) {
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

	changeAmount := inAmount - outAmount - fee
	if changeAmount > 0 {
		tx.TxOut[len(tx.TxOut)-1].Value = changeAmount
		if btcbridgetypes.IsDustOut(tx.TxOut[len(tx.TxOut)-1]) {
			tx.TxOut = tx.TxOut[0 : len(tx.TxOut)-1]
			changeAmount = 0
		}
	} else {
		tx.TxOut = tx.TxOut[0 : len(tx.TxOut)-1]

		if changeAmount < 0 {
			return nil, 0, ErrInsufficientUTXOs
		}

		changeAmount = 0
	}

	if fee < btcbridgetypes.GetTxVirtualSize(tx, utxos) {
		return nil, 0, errorsmod.Wrap(ErrFailedToBuildTx, "too low fee rate")
	}

	if err := btcbridgetypes.CheckTransactionWeight(tx, utxos); err != nil {
		return nil, 0, err
	}

	return tx, changeAmount, nil
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
	txIn.Sequence = DefaultSequence

	tx.AddTxIn(txIn)
}

// IsDustOut returns true if the given output is dust, false otherwise
// Assume that the given address is valid
func IsDustOut(value int64, address string) bool {
	addr, _ := btcutil.DecodeAddress(address, bitcoin.Network)
	pkScript, _ := txscript.PayToAddrScript(addr)

	out := wire.NewTxOut(value, pkScript)

	return !btcbridgetypes.IsOpReturnOutput(out) && mempool.IsDust(out, btcbridgetypes.MinRelayFee)
}

// IsValidBtcAddress returns true if the given address is a standard bitcoin address, false otherwise
func IsValidBtcAddress(address string) bool {
	_, err := btcutil.DecodeAddress(address, bitcoin.Network)
	return err == nil
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
