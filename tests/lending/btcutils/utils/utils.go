package utils

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"lending-tests/btcutils/types"
)

// line feed
var LineFeed = []byte{10}

// DSS file name to be excluded
const DSSFileName = ".DS_Store"

// GetContents gets the contents separated by the line feed from the given file path.
// Empty lines ignored
func GetContents(path string) (contents [][]byte, err error) {
	bz, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	for _, line := range bytes.Split(bz, LineFeed) {
		if len(line) != 0 {
			contents = append(contents, line)
		}
	}

	return contents, nil
}

// GetFiles gets files from the given path.
// Empty files ingored and subdirectories included
func GetFiles(path string) (files [][]byte, err error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	files = make([][]byte, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			if entry.Name() != DSSFileName {
				file, err := os.ReadFile(fmt.Sprintf("%s/%s", path, entry.Name()))
				if err != nil {
					return nil, err
				}

				if len(file) != 0 {
					files = append(files, file)
				}
			}
		} else {
			subDirFiles, err := GetFiles(fmt.Sprintf("%s/%s", path, entry.Name()))
			if err != nil {
				return nil, err
			}

			if len(subDirFiles) != 0 {
				files = append(files, subDirFiles...)
			}
		}
	}

	return files, nil
}

// GetSubFilePaths gets the file paths located in the given directory
func GetSubFilePaths(path string) (paths []string, err error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	paths = make([]string, len(entries))

	for _, entry := range entries {
		paths = append(paths, entry.Name())
	}

	return paths, nil
}

// WriteFile writes the specified contents to the given file
func WriteFile(path string, contents []string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, content := range contents {
		f.WriteString(fmt.Sprintf("%s\n", content))
	}

	f.WriteString("\n")

	return nil
}

// IsSegWitAddress checks if the given address is the segwit type
func IsSegWitAddress(address btcutil.Address) bool {
	switch address.(type) {
	case *btcutil.AddressWitnessPubKeyHash:
		return true

	case *btcutil.AddressWitnessScriptHash:
		return true

	case *btcutil.AddressTaproot:
		return true

	default:
		return false
	}
}

// IsTaprootAddress checks if the given address is the taproot type
func IsTaprootAddress(address btcutil.Address) bool {
	switch address.(type) {
	case *btcutil.AddressTaproot:
		return true

	default:
		return false
	}
}

// IsP2SHAddress checks if the given address is the P2SH type
func IsP2SHAddress(address btcutil.Address) bool {
	switch address.(type) {
	case *btcutil.AddressScriptHash:
		return true

	default:
		return false
	}
}

// GetRedeemScriptForNestedSegWit gets the redeem script for the P2SH-P2WPKH address
func GetRedeemScriptForNestedSegWit(pubKey string, netParams *chaincfg.Params) ([]byte, error) {
	if len(pubKey) == 0 {
		return nil, errors.New("empty public key")
	}

	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}

	p2wpkh, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pubKeyBytes), netParams)
	if err != nil {
		return nil, err
	}

	return txscript.PayToAddrScript(p2wpkh)
}

// GetXOnlyPubKey returns the X-only public key
func GetXOnlyPubKey(pubKey string) ([]byte, error) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}

	switch len(pubKeyBytes) {
	case schnorr.PubKeyBytesLen:
		return pubKeyBytes, nil

	case secp256k1.PubKeyBytesLenCompressed, secp256k1.PubKeyBytesLenUncompressed:
		return pubKeyBytes[1:33], nil

	default:
		return nil, fmt.Errorf("invalid public key length: %d", len(pubKeyBytes))
	}
}

// GetTxVirtualSize gets the virtual size of the given tx.
// Assume that tx.TxIn corresponds to the given utxos if the tx is unsigned
func GetTxVirtualSize(tx *wire.MsgTx, utxos []*types.UTXO, signed bool) int64 {
	if signed {
		return mempool.GetTxVirtualSize(btcutil.NewTx(tx))
	}

	newTx := tx.Copy()

	for i, txIn := range newTx.TxIn {
		var dummySigScript []byte
		var dummyWitness []byte

		switch txscript.GetScriptClass(utxos[i].PkScript) {
		case txscript.WitnessV1TaprootTy:
			dummyWitness = make([]byte, types.P2TRWitnessSize)

		case txscript.WitnessV0PubKeyHashTy:
			dummyWitness = make([]byte, types.P2WPKHWitnessSize)

		case txscript.ScriptHashTy:
			dummySigScript = make([]byte, types.NestedSegWitSigScriptSize)
			dummyWitness = make([]byte, types.P2WPKHWitnessSize)

		case txscript.PubKeyHashTy:
			dummySigScript = make([]byte, types.P2PKHSigScriptSize)

		default:
		}

		txIn.SignatureScript = dummySigScript
		txIn.Witness = wire.TxWitness{dummyWitness}
	}

	return mempool.GetTxVirtualSize(btcutil.NewTx(newTx))
}

// GetTxVirtualSizeFromPsbt gets the tx virtual size from the given psbt
func GetTxVirtualSizeFromPsbt(psbt *psbt.Packet) int64 {
	tx := psbt.UnsignedTx.Copy()

	for i, txIn := range tx.TxIn {
		var dummySigScript []byte
		var dummyWitness []byte

		switch txscript.GetScriptClass(GetPkScriptFromPsbt(psbt, i)) {
		case txscript.WitnessV1TaprootTy:
			dummyWitness = make([]byte, types.P2TRWitnessSize)

		case txscript.WitnessV0PubKeyHashTy:
			dummyWitness = make([]byte, types.P2WPKHWitnessSize)

		case txscript.ScriptHashTy:
			dummySigScript = make([]byte, types.NestedSegWitSigScriptSize)
			dummyWitness = make([]byte, types.P2WPKHWitnessSize)

		case txscript.PubKeyHashTy:
			dummySigScript = make([]byte, types.P2PKHSigScriptSize)

		default:
		}

		txIn.SignatureScript = dummySigScript
		txIn.Witness = wire.TxWitness{dummyWitness}
	}

	return mempool.GetTxVirtualSize(btcutil.NewTx(tx))
}

// IsDust checks if the given output is dust against the minimum relay tx fee and net params
func IsDust(txOut *wire.TxOut, minRelayTxFee int64, netParams *chaincfg.Params) bool {
	if netParams.RelayNonStdTxs || txscript.IsUnspendable(txOut.PkScript) {
		return false
	}

	return mempool.IsDust(txOut, btcutil.Amount(minRelayTxFee))
}

// GetPkScriptFromPsbt gets the pk script from the given psbt input
// Assume that the given index is valid
func GetPkScriptFromPsbt(psbt *psbt.Packet, idx int) []byte {
	witnessUtxo := psbt.Inputs[idx].WitnessUtxo
	if witnessUtxo != nil {
		return witnessUtxo.PkScript
	}

	nonWitnessUtxo := psbt.Inputs[idx].NonWitnessUtxo
	prevOutPoint := psbt.UnsignedTx.TxIn[idx].PreviousOutPoint

	return nonWitnessUtxo.TxOut[prevOutPoint.Index].PkScript
}

// AddUtxoToTx adds the given utxo to the specified tx
func AddUtxoToTx(tx *wire.MsgTx, utxo *types.UTXO) {
	txIn := new(wire.TxIn)

	txIn.PreviousOutPoint = *utxo.GetOutPoint()
	txIn.Sequence = wire.MaxTxInSequenceNum - 10

	tx.AddTxIn(txIn)
}

// AddUtxosToTx adds the given utxos to the specified tx
func AddUtxosToTx(tx *wire.MsgTx, utxos []*types.UTXO) {
	for _, utxo := range utxos {
		AddUtxoToTx(tx, utxo)
	}
}
