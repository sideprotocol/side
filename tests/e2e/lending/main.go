package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/bitcoin"
	"github.com/sideprotocol/side/crypto/adaptor"
	"github.com/sideprotocol/side/crypto/hash"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	lendingtypes "github.com/sideprotocol/side/x/lending/types"

	psbtbuilder "lending-tests/btcutils/psbt"
)

var (
	gRPC = "localhost:9090"

	binary       = path.Join(getHomeDir(), "go/bin/sided")
	globalTxArgs = "--from test --keyring-backend test --fees 1000uside --gas auto --chain-id devnet -y"

	chainParams = chaincfg.SigNetParams
)

func init() {
	bitcoin.Network = &chainParams
}

func main() {
	mode := flag.Int("mode", 1, "Specify the testing mode, 1 for liquidation, 2 for repayment")
	flag.Parse()

	if *mode == 1 {
		fmt.Printf("****testing mode: liquidation****\n\n")
	} else {
		fmt.Printf("****testing mode: repayment****\n\n")
	}

	borrowerPrivKeyHex := "7b769bcd5372539ce9ad7d4d80deb668cd07b9e6d90a6744ea7390b6b18aa55e"

	depositAmount := sdk.NewInt64Coin("sat", 100000)
	borrowAmount := sdk.NewInt64Coin("uusdc", 8000000)
	loanPeriod := 1000000 * time.Minute

	loanSecret := generateRandSecret()
	loanSecretHash := hash.Sha256(loanSecret)
	maturityTime := time.Now().Add(loanPeriod).Unix()
	finalTimeout := maturityTime + int64(7*24*time.Hour)

	fmt.Printf("loan secret: %s\n", hex.EncodeToString(loanSecret))

	borrowPrivKeyBytes, err := hex.DecodeString(borrowerPrivKeyHex)
	if err != nil {
		fmt.Printf("invalid private key\n")
		return
	}

	lendingPool, err := GetPool(gRPC)
	if err != nil {
		fmt.Printf("failed to get lending pool: %v\n", err)
		return
	}

	fmt.Printf("pool id: %s\n", lendingPool.Id)

	agency, err := GetAgency(gRPC)
	if err != nil {
		fmt.Printf("failed to get agency: %v\n", err)
		return
	}

	fmt.Printf("agency pub key: %s\n", agency.Pubkey)

	liquidationEvent, err := QueryLiquidationEvent(gRPC, depositAmount, borrowAmount)
	if err != nil {
		fmt.Printf("failed to query liquidation event: %v\n", err)
		return
	}

	fmt.Printf("oracle pub key: %s\n", liquidationEvent.OraclePubkey)
	fmt.Printf("nonce: %s\n", liquidationEvent.Nonce)
	fmt.Printf("trigger price: %s\n", liquidationEvent.Price)

	borrowerPrivKey, borrowerPubKey := btcec.PrivKeyFromBytes(borrowPrivKeyBytes)
	borrowerPubKeyHex := hex.EncodeToString(schnorr.SerializePubKey(borrowerPubKey))

	agencyPkScript, err := lendingtypes.GetPkScriptFromPubKey(agency.Pubkey)
	if err != nil {
		fmt.Printf("failed to get agency pk script: %v\n", err)
		return
	}

	taprootOutKey := txscript.ComputeTaprootKeyNoScript(borrowerPubKey)
	borrowerAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(taprootOutKey), &chaincfg.SigNetParams)
	if err != nil {
		fmt.Printf("failed to get borrower address: %v\n", err)
		return
	}

	fmt.Printf("borrower address: %s\n", borrowerAddress.EncodeAddress())

	multisigScript, err := lendingtypes.CreateMultisigScript([]string{hex.EncodeToString(schnorr.SerializePubKey(borrowerPubKey)), agency.Pubkey})
	if err != nil {
		fmt.Printf("failed to get multisig script: %v\n", err)
		return
	}

	forcedRepaymentScript, err := lendingtypes.CreateHashTimeLockScript(agency.Pubkey, hex.EncodeToString(loanSecretHash), maturityTime)
	if err != nil {
		fmt.Printf("failed to get forced repayment script: %v\n", err)
		return
	}

	timeoutRefundScript, err := lendingtypes.CreatePubKeyTimeLockScript(hex.EncodeToString(schnorr.SerializePubKey(borrowerPubKey)), int64(finalTimeout))
	if err != nil {
		fmt.Printf("failed to get timeout refund script: %v\n", err)
		return
	}

	vaultAddress, err := lendingtypes.CreateTaprootAddress(lendingtypes.GetInternalKey(), [][]byte{
		multisigScript, forcedRepaymentScript, timeoutRefundScript,
	}, &chaincfg.SigNetParams)
	if err != nil {
		fmt.Printf("failed to create vault address: %v\n", err)
		return
	}

	fmt.Printf("vault: %s\n", vaultAddress)

	// depositTxPsbt, err := psbtbuilder.BuildPsbt(borrowerAddress.EncodeAddress(), "", vaultAddress, depositAmount, 10)
	// if err != nil {
	// 	panic(err)
	// }
	depositTxPsbt, err := buildMockPsbt(vaultAddress, depositAmount.Amount.Int64())
	if err != nil {
		fmt.Printf("failed to build deposit tx psbt: %v\n", err)
		return
	}

	depositTxPsbtB64, err := depositTxPsbt.B64Encode()
	if err != nil {
		fmt.Printf("failed to serialize deposit tx psbt: %v\n", err)
		return
	}

	depositTxHash := depositTxPsbt.UnsignedTx.TxHash()

	fmt.Printf("deposit tx hash: %s\n", depositTxHash.String())
	fmt.Printf("deposit tx psbt: %s\n", depositTxPsbtB64)

	indices := make([]int, 0)
	for i := range depositTxPsbt.Inputs {
		indices = append(indices, i)
	}

	if err := psbtbuilder.SignPsbt(depositTxPsbt, indices, borrowerPrivKey, true); err != nil {
		fmt.Printf("failed to sign deposit tx: %v\n", err)
		return
	}

	signedDepositTx, err := psbt.Extract(depositTxPsbt)
	if err != nil {
		fmt.Printf("failed to extract signed deposit tx: %v\n", err)
		return
	}

	// mempoolClient := mempool.NewClient(&chaincfg.SigNetParams, base.NewClient(5, time.Second))
	// if _, err := mempoolClient.BroadcastTx(signedDepositTx); err != nil {
	// 	panic(err)
	// }

	outIndex := 0
	out := signedDepositTx.TxOut[outIndex]

	liquidationCet := wire.NewMsgTx(2)
	liquidationCet.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&depositTxHash, uint32(outIndex)), nil, nil))
	liquidationCet.AddTxOut(wire.NewTxOut(out.Value-1000, agencyPkScript))

	p, err := psbt.NewFromUnsignedTx(liquidationCet)
	if err != nil {
		fmt.Printf("failed to create liquidation cet: %v\n", err)
		return
	}

	p.Inputs[0].WitnessUtxo = signedDepositTx.TxOut[outIndex]
	p.Inputs[0].SighashType = txscript.SigHashDefault

	liquidationCetPsbt, err := p.B64Encode()
	if err != nil {
		fmt.Printf("failed to serialize liquidation cet psbt: %v\n", err)
		return
	}

	fmt.Printf("liquidation cet: %s\n", liquidationCetPsbt)

	sigHash, err := lendingtypes.CalcTapscriptSigHash(p, 0, txscript.SigHashDefault, multisigScript)
	if err != nil {
		fmt.Printf("failed to calculate sig hash: %v\n", err)
		return
	}

	signaturePoint, _ := hex.DecodeString(liquidationEvent.SignaturePoint)

	adaptorSig, err := adaptor.Sign(borrowerPrivKey, sigHash, signaturePoint)
	if err != nil {
		fmt.Printf("failed to get the adaptor signature: %v\n", err)
		return
	}

	fmt.Printf("adaptor signature: %s\n", hex.EncodeToString(adaptorSig.Serialize()))
	fmt.Printf("adaptor signature verified: %t\n", adaptor.Verify(adaptorSig.Serialize(), sigHash, schnorr.SerializePubKey(borrowerPubKey), signaturePoint))

	applyTxArgs := fmt.Sprintf("tx lending apply %s %s %d %d %s %s %s %d %d %s %s %s", borrowerPubKeyHex, hex.EncodeToString(loanSecretHash), maturityTime, finalTimeout, depositTxPsbtB64, lendingPool.Id, borrowAmount.String(), liquidationEvent.EventId, agency.Id, liquidationCetPsbt, hex.EncodeToString(adaptorSig.Serialize()), globalTxArgs)
	approveTxArgs := fmt.Sprintf("tx lending approve %s %s %s %s", depositTxHash.String(), "4fc4af9a4fac617aa4d7152313c56678469c93ad4f07b4864d77295bee9d79e8", "12559b5ef74508404ba567b4499c58f9e0ab5c7d34257276f37cdc810d441c00", globalTxArgs)
	redeemTxArgs := fmt.Sprintf("tx lending redeem %s %s %s", vaultAddress, hex.EncodeToString(loanSecret), globalTxArgs)

	if err := Apply(binary, applyTxArgs); err != nil {
		fmt.Printf("failed to execute apply tx: %v\n", err)
		return
	}

	time.Sleep(10 * time.Second)

	if err := Approve(binary, approveTxArgs); err != nil {
		fmt.Printf("failed to execute approve tx: %v\n", err)
		return
	}

	time.Sleep(10 * time.Second)

	if err := Redeem(binary, redeemTxArgs); err != nil {
		fmt.Printf("failed to execute redeem tx: %v\n", err)
		return
	}

	time.Sleep(10 * time.Second)

	switch *mode {
	case 1:
		// set price to liquidate the loan

		triggerPrice, _ := strconv.ParseUint(liquidationEvent.Price, 10, 64)

		setPriceTxArgs := fmt.Sprintf("tx lending submit-price %d %s", triggerPrice-100, globalTxArgs)

		if err := SetPrice(binary, setPriceTxArgs); err != nil {
			fmt.Printf("failed to execute set price tx: %v\n", err)
			return
		}
	case 2:
		// repay the loan

		adaptorSecret := generateAdaptorSecret()
		adaptorPoint := hex.EncodeToString(adaptorSecret.PubKey().SerializeCompressed())

		fmt.Printf("repayment adaptor secret: %s\n", hex.EncodeToString(adaptorSecret.Serialize()))
		fmt.Printf("repayment adaptor point: %s\n", adaptorPoint)

		repayTxArgs := fmt.Sprintf("tx lending repay %s %s %s", vaultAddress, adaptorPoint, globalTxArgs)

		if err := Repay(binary, repayTxArgs); err != nil {
			fmt.Printf("failed to execute repay tx: %v\n", err)
			return
		}

		time.Sleep(10 * time.Second)

		var repayment *lendingtypes.Repayment

		// query the repayment
		for {
			repayment, err = GetRepayment(gRPC, vaultAddress)
			if err != nil {
				fmt.Printf("failed to query repayment: %v\n", err)

				time.Sleep(2 * time.Second)
				continue
			}

			if len(repayment.RepayAdaptorPoint) == 0 {
				fmt.Printf("no repayment adaptor point found yet, waiting or repay manually if failed")

				time.Sleep(5 * time.Second)
				continue
			}

			if len(repayment.DcaAdaptorSignatures) == 0 {
				fmt.Printf("no dca adaptor signature found yet, waiting")

				time.Sleep(5 * time.Second)
				continue
			}

			// dca adaptor signatures submitted
			break
		}

		// decrypt adaptor signatures
		agencySigs, err := decryptAdaptorSignatures(adaptorSecret, repayment.DcaAdaptorSignatures)
		if err != nil {
			fmt.Printf("failed to decrypt dca adaptor signatures: %v\n", err)
			return
		}

		// build signed repayment tx
		signedTx, err := buildSignedRepaymentTx(repayment.Tx, multisigScript, agencySigs, borrowerPrivKey)
		if err != nil {
			fmt.Printf("failed to build signed repayment tx: %v\n", err)
			return
		}

		// send the signed tx to the Bitcoin network
		// if _, err := mempoolClient.BroadcastTx(signedTx); err != nil {
		// 	fmt.Printf("failed to broadcast tx: %v\n", err)
		// 	return
		// }

		signedTxBytes, err := serializeTx(signedTx)
		if err != nil {
			fmt.Printf("failed to serialize tx: %v\n", err)
			return
		}

		fmt.Printf("signed repayment tx: %s\n", hex.EncodeToString(signedTxBytes))

		// close loan with the first borrower signature
		borrowerSig := signedTx.TxIn[0].Witness[1]

		closeTxArgs := fmt.Sprintf("tx lending close %s %s %s", vaultAddress, hex.EncodeToString(borrowerSig), globalTxArgs)

		if err := Close(binary, closeTxArgs); err != nil {
			fmt.Printf("failed to execute close tx: %v\n", err)
			return
		}
	}

	fmt.Printf("operations finished\n")
}

func GetPool(gRPC string) (*lendingtypes.LendingPool, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := lendingtypes.NewQueryClient(conn)

	resp, err := client.Pools(context.Background(), &lendingtypes.QueryPoolsRequest{})
	if err != nil {
		return nil, err
	}

	if len(resp.Pools) == 0 {
		return nil, fmt.Errorf("no pool created yet")
	}

	return resp.Pools[0], nil
}

func GetAgency(gRPC string) (*dlctypes.Agency, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := dlctypes.NewQueryClient(conn)

	resp, err := client.Agencies(context.Background(), &dlctypes.QueryAgenciesRequest{
		Status: dlctypes.AgencyStatus_Agency_status_Enable,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Agencies) == 0 {
		return nil, fmt.Errorf("no enabled agencies")
	}

	return resp.Agencies[0], nil
}

func QueryLiquidationEvent(gRPC string, collateralAmount sdk.Coin, borrowAmount sdk.Coin) (*lendingtypes.QueryLiquidationEventResponse, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := lendingtypes.NewQueryClient(conn)

	resp, err := client.LiquidationEvent(context.Background(), &lendingtypes.QueryLiquidationEventRequest{
		CollateralAcmount: &collateralAmount,
		BorrowAmount:      &borrowAmount,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetRepayment(gRPC string, loanId string) (*lendingtypes.Repayment, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := lendingtypes.NewQueryClient(conn)

	resp, err := client.Repayment(context.Background(), &lendingtypes.QueryRepaymentRequest{
		LoanId: loanId,
	})
	if err != nil {
		return nil, err
	}

	return resp.Repayment, nil
}

func Apply(binary string, args string) error {
	fmt.Printf("execute apply tx: \n\n")

	return executeCmd(binary, args)
}

func Approve(binary string, args string) error {
	fmt.Printf("execute approve tx: \n\n")

	return executeCmd(binary, args)
}

func Redeem(binary string, args string) error {
	fmt.Printf("execute redeem tx: \n\n")

	return executeCmd(binary, args)
}

func SetPrice(binary string, args string) error {
	fmt.Printf("execute submit price tx: \n\n")

	return executeCmd(binary, args)
}

func Repay(binary string, args string) error {
	fmt.Printf("execute repay tx: \n\n")

	return executeCmd(binary, args)
}

func Close(binary string, args string) error {
	fmt.Printf("execute close tx: \n\n")

	return executeCmd(binary, args)
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

func buildMockPsbt(recipient string, amount int64) (*psbt.Packet, error) {
	recipientPkScript, err := lendingtypes.GetPkScriptFromAddress(recipient)
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(2)

	txIn := wire.NewTxIn(wire.NewOutPoint((*chainhash.Hash)(chainhash.HashB([]byte{})), 0), nil, nil)
	txOut := wire.NewTxOut(amount, recipientPkScript)

	tx.AddTxIn(txIn)
	tx.AddTxOut(txOut)

	p, err := psbt.NewFromUnsignedTx(tx)
	if err != nil {
		return nil, err
	}

	for i := range p.Inputs {
		p.Inputs[i].SighashType = txscript.SigHashDefault
		p.Inputs[i].WitnessUtxo = txOut
	}

	return p, nil
}

func buildSignedRepaymentTx(repaymentTx string, script []byte, agencySigs [][]byte, privKey *secp256k1.PrivateKey) (*wire.MsgTx, error) {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(repaymentTx)), true)
	if err != nil {
		return nil, err
	}

	for i, input := range p.Inputs {
		selfSig, err := lendingtypes.CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return nil, err
		}

		p.Inputs[i].TaprootScriptSpendSig = []*psbt.TaprootScriptSpendSig{
			{
				Signature: agencySigs[i],
			},
			{
				Signature: selfSig,
			},
		}
	}

	if err := psbt.MaybeFinalizeAll(p); err != nil {
		return nil, err
	}

	signedTx, err := psbt.Extract(p)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func generateRandSecret() []byte {
	secret := make([]byte, 32)
	rand.Read(secret)

	return secret
}

func generateAdaptorSecret() *secp256k1.PrivateKey {
	secretKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}

	return secretKey
}

func decryptAdaptorSignatures(adaptorSecret *secp256k1.PrivateKey, adaptorSigs []string) ([][]byte, error) {
	adaptedSigs := make([][]byte, 0)

	for _, adaptorSig := range adaptorSigs {
		sigBytes, err := hex.DecodeString(adaptorSig)
		if err != nil {
			return nil, err
		}

		adaptedSig := adaptor.Adapt(sigBytes, adaptorSecret.Serialize())
		adaptedSigs = append(adaptedSigs, adaptedSig)
	}

	return adaptedSigs, nil
}

func serializeTx(tx *wire.MsgTx) ([]byte, error) {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return home
}

func executeCmd(name string, args string) error {
	cmd := exec.Command(name, strings.Split(args, " ")...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
