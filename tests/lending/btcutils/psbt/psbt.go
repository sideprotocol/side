package psbt

import (
	"fmt"
	"sort"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"lending-tests/btcutils/client/base"
	"lending-tests/btcutils/client/btcapi/mempool"
	"lending-tests/btcutils/types"
	"lending-tests/btcutils/utils"
)

func BuildPsbt(sender, pubKey, recipient string, amount, feeRate int64) (*psbt.Packet, error) {
	senderAddress, err := btcutil.DecodeAddress(sender, &chaincfg.SigNetParams)
	if err != nil {
		return nil, err
	}

	recipientAddress, err := btcutil.DecodeAddress(recipient, &chaincfg.SigNetParams)
	if err != nil {
		return nil, err
	}

	recipientPkScript, err := txscript.PayToAddrScript(recipientAddress)
	if err != nil {
		return nil, err
	}

	senderUtxos, err := GetUtxos(senderAddress)
	if err != nil {
		return nil, err
	}

	txOuts := make([]*wire.TxOut, 0)
	txOuts = append(txOuts, wire.NewTxOut(amount, recipientPkScript))

	unsignedTx, paymentUtxos, err := BuildTransaction(nil, txOuts, senderUtxos, senderAddress, feeRate)
	if err != nil {
		return nil, err
	}

	p, err := psbt.NewFromUnsignedTx(unsignedTx)
	if err != nil {
		return nil, err
	}

	for i, utxo := range paymentUtxos {
		if err := AddInputToPsbt(p, i, utxo, senderAddress, pubKey, txscript.SigHashAll); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func SignPsbt(p *psbt.Packet, inIndice []int, key *secp256k1.PrivateKey, finalize bool) error {
	prevOutputFetcher := txscript.NewMultiPrevOutFetcher(nil)

	for i, txIn := range p.UnsignedTx.TxIn {
		prevOutput := p.Inputs[i].WitnessUtxo
		if prevOutput == nil {
			prevOutput = p.Inputs[i].NonWitnessUtxo.TxOut[txIn.PreviousOutPoint.Index]
		}

		prevOutputFetcher.AddPrevOut(txIn.PreviousOutPoint, prevOutput)
	}

	for _, idx := range inIndice {
		output := p.Inputs[idx].WitnessUtxo
		hashType := p.Inputs[idx].SighashType

		witness, err := txscript.TaprootWitnessSignature(p.UnsignedTx, txscript.NewTxSigHashes(p.UnsignedTx, prevOutputFetcher),
			idx, output.Value, output.PkScript, hashType, key)
		if err != nil {
			return err
		}

		p.Inputs[idx].TaprootKeySpendSig = witness[0]

		if finalize {
			if err := psbt.Finalize(p, idx); err != nil {
				return err
			}
		}
	}

	return nil
}

func GetSignedTx(p *psbt.Packet) (*wire.MsgTx, error) {
	if err := psbt.MaybeFinalizeAll(p); err != nil {
		return nil, err
	}

	return psbt.Extract(p)
}

// BuildTransaction builds an unsigned tx from the given params.
func BuildTransaction(utxos []*types.UTXO, txOuts []*wire.TxOut, paymentUtxos []*types.UTXO, changeAddress btcutil.Address, feeRate int64) (*wire.MsgTx, []*types.UTXO, error) {
	tx := wire.NewMsgTx(types.TxVersion)

	inAmount := int64(0)
	outAmount := int64(0)

	for _, utxo := range utxos {
		utils.AddUtxoToTx(tx, utxo)
		inAmount += utxo.Value
	}

	for _, txOut := range txOuts {
		if utils.IsDust(txOut, types.MinRelayTxFee, &chaincfg.SigNetParams) {
			return nil, nil, fmt.Errorf("dust output value: %d", txOut.Value)
		}

		tx.AddTxOut(txOut)
		outAmount += txOut.Value
	}

	changePkScript, err := txscript.PayToAddrScript(changeAddress)
	if err != nil {
		return nil, nil, err
	}

	changeOut := wire.NewTxOut(0, changePkScript)

	selectedPaymentUtxos, err := AddPaymentUtxosToTx(tx, utxos, inAmount-outAmount, paymentUtxos, changeOut, feeRate)
	if err != nil {
		return nil, nil, err
	}

	return tx, selectedPaymentUtxos, nil
}

// AddPaymentUtxosToTx adds the payment utxos to the tx
func AddPaymentUtxosToTx(tx *wire.MsgTx, utxos []*types.UTXO, inOutdiff int64, paymentUtxos []*types.UTXO, changeOut *wire.TxOut, feeRate int64) ([]*types.UTXO, error) {
	selectedPaymentUtxos := make([]*types.UTXO, 0)
	paymentValue := int64(0)

	sort.Slice(paymentUtxos, func(i, j int) bool {
		return paymentUtxos[i].Value > paymentUtxos[j].Value
	})

	for _, utxo := range paymentUtxos {
		utils.AddUtxoToTx(tx, utxo)
		tx.AddTxOut(changeOut)

		utxos = append(utxos, utxo)
		selectedPaymentUtxos = append(selectedPaymentUtxos, utxo)

		paymentValue += utxo.Value
		fee := utils.GetTxVirtualSize(tx, utxos, false) * feeRate

		changeValue := paymentValue + inOutdiff - fee
		if changeValue > 0 {
			tx.TxOut[len(tx.TxOut)-1].Value = changeValue
			if utils.IsDust(tx.TxOut[len(tx.TxOut)-1], types.MinRelayTxFee, &chaincfg.SigNetParams) {
				tx.TxOut = tx.TxOut[0 : len(tx.TxOut)-1]
			}

			return selectedPaymentUtxos, nil
		} else {
			tx.TxOut = tx.TxOut[0 : len(tx.TxOut)-1]

			if changeValue == 0 {
				return selectedPaymentUtxos, nil
			}

			if changeValue < 0 {
				feeWithoutChange := utils.GetTxVirtualSize(tx, utxos, false) * feeRate
				if paymentValue+inOutdiff-feeWithoutChange >= 0 {
					return selectedPaymentUtxos, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("insufficient utxos")
}

// AddInputToPsbt adds the given utxo to the psbt.
// Assume that the input index is valid
func AddInputToPsbt(p *psbt.Packet, index int, utxo *types.UTXO, address btcutil.Address, pubKey string, sigHashType txscript.SigHashType) error {
	p.Inputs[index].SighashType = sigHashType
	p.Inputs[index].WitnessUtxo = utxo.GetOutput()

	return nil
}

// GetUtxos retrieves the utxos of the given address
func GetUtxos(address btcutil.Address) ([]*types.UTXO, error) {
	mempoolClient := mempool.NewClient(&chaincfg.SigNetParams, base.NewClient(5, time.Second))

	utxos := make([]*types.UTXO, 0)

	unspentList, err := mempoolClient.ListUnspent(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get utxos, err: %v", err)
	}

	for i := range unspentList {
		utxos = append(utxos, types.NewUTXO(
			&unspentList[i].Outpoint.Hash,
			unspentList[i].Outpoint.Index,
			unspentList[i].Output.Value,
			unspentList[i].Output.PkScript,
		))
	}

	return utxos, nil
}
