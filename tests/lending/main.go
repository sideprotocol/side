package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/adaptor"
	"github.com/sideprotocol/side/crypto/hash"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	lendingtypes "github.com/sideprotocol/side/x/lending/types"

	"lending-tests/btcutils/client/base"
	"lending-tests/btcutils/client/btcapi/mempool"
	psbtbuilder "lending-tests/btcutils/psbt"
)

var chainParams = chaincfg.SigNetParams

func init() {
	config := sdk.GetConfig()
	config.SetBtcChainCfg(&chaincfg.SigNetParams)
}

func main() {
	borrowerPrivKeyHex := "7b769bcd5372539ce9ad7d4d80deb668cd07b9e6d90a6744ea7390b6b18aa55e"
	agencyPubKeyHex := "8f05f98ecdc52ff4019fd904869e539f35b847698aab2acb257e26dee539a45b"

	depositAmount := int64(100000)  // 0.001 btc
	borrowAmount := int64(70000000) // 70 usdc
	loanSecret := hash.Sha256([]byte("secret"))
	loanSecretHash := hash.Sha256(loanSecret)
	lockTime := int64(100000)
	finalTimeout := int64(1000)

	oraclePubKey := "8f05f98ecdc52ff4019fd904869e539f35b847698aab2acb257e26dee539a45b"
	nonce := "8f05f98ecdc52ff4019fd904869e539f35b847698aab2acb257e26dee539a45b"
	triggerPrice := lendingtypes.GetLiquidationPrice(sdkmath.NewInt(depositAmount), sdkmath.NewInt(borrowAmount), sdkmath.NewInt(70))

	println("trigger price:", triggerPrice.String())

	borrowPrivKeyBytes, err := hex.DecodeString(borrowerPrivKeyHex)
	if err != nil {
		panic(err)
	}

	borrowerPrivKey, borrowerPubKey := btcec.PrivKeyFromBytes(borrowPrivKeyBytes)

	agencyPkScript, err := lendingtypes.GetPkScriptFromPubKey(agencyPubKeyHex)
	if err != nil {
		panic(err)
	}

	taprootOutKey := txscript.ComputeTaprootKeyNoScript(borrowerPubKey)
	borrowerAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(taprootOutKey), &chaincfg.SigNetParams)
	if err != nil {
		panic(err)
	}

	fmt.Println("borrower address:", borrowerAddress.EncodeAddress())

	multisigScript, err := lendingtypes.CreateMultisigScript([]string{hex.EncodeToString(borrowerPubKey.SerializeCompressed()[1:]), agencyPubKeyHex})
	if err != nil {
		panic(err)
	}

	forcedRepaymentScript, err := lendingtypes.CreateHashTimeLockScript(agencyPubKeyHex, hex.EncodeToString(loanSecretHash), lockTime)
	if err != nil {
		panic(err)
	}

	timeoutRefundScript, err := lendingtypes.CreatePubKeyTimeLockScript(hex.EncodeToString(borrowerPubKey.SerializeCompressed()[1:]), finalTimeout)
	if err != nil {
		panic(err)
	}

	vaultAddress, err := lendingtypes.CreateTaprootAddress(lendingtypes.GetInternalKey(), [][]byte{
		multisigScript, forcedRepaymentScript, timeoutRefundScript,
	}, &chaincfg.SigNetParams)
	if err != nil {
		panic(err)
	}

	println("vault:", vaultAddress)

	depositTxPsbt, err := psbtbuilder.BuildPsbt(borrowerAddress.EncodeAddress(), "", vaultAddress, int64(depositAmount), 10)
	if err != nil {
		panic(err)
	}

	indices := make([]int, 0)
	for i := range depositTxPsbt.Inputs {
		indices = append(indices, i)
	}

	if err := psbtbuilder.SignPsbt(depositTxPsbt, indices, borrowerPrivKey, true); err != nil {
		panic(err)
	}

	signedDepositTx, err := psbt.Extract(depositTxPsbt)
	if err != nil {
		panic(err)
	}

	mempoolClient := mempool.NewClient(&chaincfg.SigNetParams, base.NewClient(5, time.Second))
	depositTxHash, err := mempoolClient.BroadcastTx(signedDepositTx)
	if err != nil {
		panic(err)
	}

	fmt.Println("deposit tx hash:", depositTxHash.String())

	outIndex := 0
	out := signedDepositTx.TxOut[outIndex]

	liquidationCet := wire.NewMsgTx(2)
	liquidationCet.AddTxIn(wire.NewTxIn(wire.NewOutPoint(depositTxHash, uint32(outIndex)), nil, nil))
	liquidationCet.AddTxOut(wire.NewTxOut(out.Value-2000, agencyPkScript))

	p, err := psbt.NewFromUnsignedTx(liquidationCet)
	if err != nil {
		panic(err)
	}

	p.Inputs[0].WitnessUtxo = signedDepositTx.TxOut[outIndex]
	p.Inputs[0].SighashType = txscript.SigHashDefault

	liquidationCetPsbt, err := p.B64Encode()
	if err != nil {
		panic(err)
	}

	fmt.Println("liquidation cet:", liquidationCetPsbt)

	sigHash, err := lendingtypes.CalcTapscriptSigHash(p, 0, txscript.SigHashDefault, multisigScript)
	if err != nil {
		panic(err)
	}

	signaturePoint, err := dlctypes.GetSignaturePointFromEvent(&dlctypes.DLCPriceEvent{
		Pubkey:       oraclePubKey,
		Nonce:        nonce,
		TriggerPrice: triggerPrice,
	})
	if err != nil {
		panic(err)
	}

	adaptorSig, err := adaptor.Sign(borrowerPrivKey, sigHash, signaturePoint)
	if err != nil {
		panic(err)
	}

	fmt.Println("adaptor sig:", hex.EncodeToString(adaptorSig.Serialize()))
}

func SignTaprootTransaction(key *secp256k1.PrivateKey, tx *wire.MsgTx, prevOuts []*wire.TxOut, hashType txscript.SigHashType) error {
	prevOutFetcher := txscript.NewMultiPrevOutFetcher(nil)

	for i := range tx.TxIn {
		prevOutFetcher.AddPrevOut(tx.TxIn[i].PreviousOutPoint, prevOuts[i])
	}

	for i, txIn := range tx.TxIn {
		witness, err := txscript.TaprootWitnessSignature(tx, txscript.NewTxSigHashes(tx, prevOutFetcher), i, prevOuts[i].Value, prevOuts[i].PkScript, hashType, key)
		if err != nil {
			return err
		}

		txIn.Witness = witness
	}

	return nil
}

func SignTapscript(key *secp256k1.PrivateKey, tx *wire.MsgTx, prevOuts []*wire.TxOut, idx int, script []byte, hashType txscript.SigHashType) ([]byte, error) {
	prevOutFetcher := txscript.NewMultiPrevOutFetcher(nil)

	for i := range tx.TxIn {
		prevOutFetcher.AddPrevOut(tx.TxIn[i].PreviousOutPoint, prevOuts[i])
	}

	signature, err := txscript.RawTxInTapscriptSignature(tx, txscript.NewTxSigHashes(tx, prevOutFetcher), idx, prevOuts[idx].Value, prevOuts[idx].PkScript, txscript.NewBaseTapLeaf(script), hashType, key)
	if err != nil {
		return nil, err
	}

	return signature, nil
}
