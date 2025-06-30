package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
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
	"github.com/sideprotocol/side/bitcoin/crypto/adaptor"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	lendingtypes "github.com/sideprotocol/side/x/lending/types"

	"lending-tests/btcutils/client/base"
	"lending-tests/btcutils/client/btcapi/mempool"
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
	mode := flag.Int("mode", 2, "Specify the testing mode, 1 for liquidation(Deprecated), 2 for repayment")
	flag.Parse()

	if *mode == 1 {
		fmt.Println("****testing mode: liquidation****\n")
		fmt.Println("the mode has been deprecated")

		return
	} else {
		fmt.Println("****testing mode: repayment****\n")
	}

	poolId := "usdc"
	collateralAmount := sdk.NewInt64Coin("sat", 100000)
	borrowAmount := sdk.NewInt64Coin("uusdc", 100000000)
	maturity := 7 * 24 * 3600 // 7 days

	borrowerPrivKeyHex := "7b769bcd5372539ce9ad7d4d80deb668cd07b9e6d90a6744ea7390b6b18aa55e"

	borrowerPrivKeyBytes, err := hex.DecodeString(borrowerPrivKeyHex)
	if err != nil {
		fmt.Printf("invalid private key\n")
		return
	}

	borrowerPrivKey, borrowerPubKey := btcec.PrivKeyFromBytes(borrowerPrivKeyBytes)
	borrowerPubKeyHex := hex.EncodeToString(schnorr.SerializePubKey(borrowerPubKey))

	_, err = GetPool(gRPC, poolId)
	if err != nil {
		fmt.Printf("failed to get lending pool %s: %v\n", poolId, err)
		return
	}

	liquidationPrice, err := QueryLiquidationPrice(gRPC, poolId, collateralAmount.String(), borrowAmount.String(), int64(maturity))
	if err != nil {
		fmt.Printf("failed to query liquidation price: %v\n", err)
		return
	}

	fmt.Printf("liquidation price: %s%s\n", liquidationPrice.Price, liquidationPrice.Pair)

	dcm, err := GetDCM(gRPC)
	if err != nil {
		fmt.Printf("failed to get dcm: %v\n", err)
		return
	}

	fmt.Printf("dcm pub key: %s\n", dcm.Pubkey)

	dlcEventCount, err := QueryAvailableDLCEventCount(gRPC)
	if err != nil {
		fmt.Printf("failed to query available dlc event count: %v\n", err)
		return
	}

	if dlcEventCount.Count == 0 {
		fmt.Printf("no available dlc event\n")
		return
	}

	fmt.Printf("available dlc event count: %d\n", dlcEventCount.Count)

	applyTxArgs := fmt.Sprintf("tx lending apply %s %s %s %s %d %d %s %s %s", borrowerPubKeyHex, borrowerPubKeyHex, poolId, borrowAmount, maturity, dcm.Id, globalTxArgs)
	if err := Apply(binary, applyTxArgs); err != nil {
		fmt.Printf("failed to execute apply tx: %v\n", err)
		return
	}

	time.Sleep(10 * time.Second)

	loans, err := QueryLoans(gRPC)
	if err != nil {
		fmt.Printf("failed to query loans: %v\n", err)
		return
	}

	vault := loans.Loans[0].VaultAddress
	loanId := vault

	fmt.Printf("vault: %s\n", vault)

	// deposit btc

	taprootOutKey := txscript.ComputeTaprootKeyNoScript(borrowerPubKey)
	borrowerAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(taprootOutKey), &chaincfg.TestNet3Params)
	if err != nil {
		fmt.Printf("failed to get borrower btc address: %v\n", err)
		return
	}

	fmt.Printf("borrower btc address: %s\n", borrowerAddress.EncodeAddress())

	// depositTxPsbt, err := buildMockPsbt(vaultAddress, collateral.Amount.Int64())
	// if err != nil {
	// 	fmt.Printf("failed to build deposit tx psbt: %v\n", err)
	// 	return
	// }
	depositTxPsbt, err := psbtbuilder.BuildPsbt(borrowerAddress.EncodeAddress(), "", vault, collateralAmount.Amount.Int64(), 10)
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

	mempoolClient := mempool.NewClient(&chaincfg.TestNet3Params, base.NewClient(5, time.Second))
	if _, err := mempoolClient.BroadcastTx(signedDepositTx); err != nil {
		fmt.Printf("failed to broadcast deposit tx: %v\n", err)
		return
	}

	// query cet signing infos

	cetInfos, err := QueryCetInfos(gRPC, loanId)
	if err != nil {
		fmt.Printf("failed to query cet infos: %v\n", err)
		return
	}

	fmt.Printf("cet infos: %+v\n", cetInfos)

	// build cets

	dcmPkScript, err := lendingtypes.GetPkScriptFromPubKey(dcm.Pubkey)
	if err != nil {
		fmt.Printf("failed to get dcm pk script: %v\n", err)
		return
	}

	borrowerPkScript, err := lendingtypes.GetPkScriptFromPubKey(borrowerPubKeyHex)
	if err != nil {
		fmt.Printf("failed to get borrower pk script: %v\n", err)
		return
	}

	liquidationCetPsbt, liquidationAdaptorSignatures, err := buildLiquidationCet(signedDepositTx, dcmPkScript, cetInfos.LiquidationCetInfo, borrowerPrivKey)
	if err != nil {
		fmt.Printf("failed to build liquidation cet: %v\n", err)
		return
	}

	_, defaultLiquidationAdaptorSignatures, err := buildLiquidationCet(signedDepositTx, dcmPkScript, cetInfos.DefaultLiquidationCetInfo, borrowerPrivKey)
	if err != nil {
		fmt.Printf("failed to build default liquidation cet: %v\n", err)
		return
	}

	repaymentCetPsbt, repaymentSignatures, err := buildRepaymentCet(signedDepositTx, borrowerPkScript, cetInfos.RepaymentCetInfo, borrowerPrivKey)
	if err != nil {
		fmt.Printf("failed to build repayment cet: %v\n", err)
		return
	}

	// submit cets (authorize)

	submitCetsTxArgs := fmt.Sprintf("tx lending submit-cets %s %s %s %s %s %s %s %s", loanId, depositTxPsbtB64, liquidationCetPsbt, liquidationAdaptorSignatures[0], defaultLiquidationAdaptorSignatures[0], repaymentCetPsbt, repaymentSignatures[0], globalTxArgs)
	if err := SubmitCets(binary, submitCetsTxArgs); err != nil {
		fmt.Printf("failed to execute authorization tx: %v\n", err)
		return
	}

	time.Sleep(10 * time.Second)

	// submit deposit tx to Side

	// TODO: retrieve tx proof from mempool
	blockHash := ""
	txProof := ""

	submitDepositTxArgs := fmt.Sprintf("tx lending submit-deposit-tx %s %s %s %s %s", vault, serializeTxB64(signedDepositTx), blockHash, txProof, globalTxArgs)
	if err := SubmitCets(binary, submitDepositTxArgs); err != nil {
		fmt.Printf("failed to execute submit deposit tx: %v\n", err)
		return
	}

	switch *mode {
	case 1:
		// set price to liquidate the loan
		// deprecated due to that the price testing method has been removed

		fmt.Printf("Deprecated mode\n")

		return

	case 2:
		// check if the loan is open

		fmt.Printf("check if the loan is open...\n")

		for {
			loan, err := QueryLoan(gRPC, loanId)
			if err != nil {
				fmt.Printf("failed to query loan %s: %v\n", loanId, err)

				time.Sleep(2 * time.Second)
				continue
			}

			if loan.Loan.Status == lendingtypes.LoanStatus_Open {
				break
			}

			time.Sleep(2 * time.Second)
		}

		// repay
		repayTxArgs := fmt.Sprintf("tx lending repay %s %s", loanId, globalTxArgs)
		if err := Repay(binary, repayTxArgs); err != nil {
			fmt.Printf("failed to execute repay tx: %v\n", err)
			return
		}

		time.Sleep(10 * time.Second)

		// query the loan status

		fmt.Printf("check if the loan is repaid...\n")

		for {
			loan, err := QueryLoan(gRPC, loanId)
			if err != nil {
				fmt.Printf("failed to query loan: %v\n", err)

				time.Sleep(2 * time.Second)
				continue
			}

			if loan.Loan.Status == lendingtypes.LoanStatus_Repaid {
				break
			}

			time.Sleep(2 * time.Second)
		}

		// query the signed repayment tx

		var rawRepaymentTx []byte

		fmt.Printf("query the signed repayment tx...\n")

		for {
			dlcMeta, err := GetDLCMeta(gRPC, loanId)
			if err != nil {
				fmt.Printf("failed to query dlc meta: %v\n", err)

				time.Sleep(2 * time.Second)
				continue
			}

			if len(dlcMeta.RepaymentCet.SignedTxHex) != 0 {
				fmt.Printf("signed repayment tx: %s\n", dlcMeta.RepaymentCet.SignedTxHex)

				rawRepaymentTx, _ = hex.DecodeString(dlcMeta.RepaymentCet.SignedTxHex)
				break
			}

			time.Sleep(2 * time.Second)
		}

		// deserialize raw repayment tx
		var signedRepaymentTx wire.MsgTx
		if err := signedRepaymentTx.Deserialize(bytes.NewReader(rawRepaymentTx)); err != nil {
			fmt.Printf("failed to deserialize repayment tx: %v\n", err)
			return
		}

		fmt.Printf("repayment tx hash: %s\n", signedRepaymentTx.TxHash().String())

		// send the signed tx to the Bitcoin network
		if _, err := mempoolClient.BroadcastTx(&signedRepaymentTx); err != nil {
			fmt.Printf("failed to broadcast repayment tx: %v\n", err)
			return
		}

		fmt.Printf("repayment tx broadcasted\n")
	}

	fmt.Printf("operations finished\n")
}

func GetPool(gRPC string, id string) (*lendingtypes.LendingPool, error) {
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

	for _, pool := range resp.Pools {
		if pool.Id == id {
			return pool, nil
		}
	}

	return nil, fmt.Errorf("pool %s not found", id)
}

func QueryLiquidationPrice(gRPC string, poolId string, collateralAmount string, borrowAmount string, maturity int64) (*lendingtypes.QueryLiquidationPriceResponse, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := lendingtypes.NewQueryClient(conn)

	resp, err := client.LiquidationPrice(context.Background(), &lendingtypes.QueryLiquidationPriceRequest{
		PoolId:           poolId,
		CollateralAmount: collateralAmount,
		BorrowAmount:     borrowAmount,
		Maturity:         maturity,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetDCM(gRPC string) (*dlctypes.DCM, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := dlctypes.NewQueryClient(conn)

	resp, err := client.DCMs(context.Background(), &dlctypes.QueryDCMsRequest{
		Status: dlctypes.DCMStatus_DCM_status_Enable,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.DCMs) == 0 {
		return nil, fmt.Errorf("no enabled dcms")
	}

	return resp.DCMs[0], nil
}

func QueryAvailableDLCEventCount(gRPC string) (*lendingtypes.QueryDlcEventCountResponse, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := lendingtypes.NewQueryClient(conn)

	resp, err := client.DlcEventCount(context.Background(), &lendingtypes.QueryDlcEventCountRequest{})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func QueryLoans(gRPC string) (*lendingtypes.QueryLoansResponse, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := lendingtypes.NewQueryClient(conn)

	resp, err := client.Loans(context.Background(), &lendingtypes.QueryLoansRequest{
		Status: lendingtypes.LoanStatus_Requested,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func QueryLoan(gRPC string, id string) (*lendingtypes.QueryLoanResponse, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := lendingtypes.NewQueryClient(conn)

	resp, err := client.Loan(context.Background(), &lendingtypes.QueryLoanRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func QueryCetInfos(gRPC string, loanId string) (*lendingtypes.QueryLoanCetInfosResponse, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := lendingtypes.NewQueryClient(conn)

	resp, err := client.LoanCetInfos(context.Background(), &lendingtypes.QueryLoanCetInfosRequest{
		LoanId: loanId,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetDLCMeta(gRPC string, loanId string) (*lendingtypes.DLCMeta, error) {
	conn, err := grpc.NewClient(gRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := lendingtypes.NewQueryClient(conn)

	resp, err := client.LoanDlcMeta(context.Background(), &lendingtypes.QueryLoanDlcMetaRequest{
		LoanId: loanId,
	})
	if err != nil {
		return nil, err
	}

	return resp.DlcMeta, nil
}

func Apply(binary string, args string) error {
	fmt.Printf("execute apply tx: \n\n")

	return executeCmd(binary, args)
}

func SubmitDepositTx(binary string, args string) error {
	fmt.Printf("execute submit deposit tx: \n\n")

	return executeCmd(binary, args)
}

func SubmitCets(binary string, args string) error {
	fmt.Printf("execute submit cets tx: \n\n")

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

func buildLiquidationCet(depositTx *wire.MsgTx, dcmPkScript []byte, cetInfo *lendingtypes.CetInfo, privKey *secp256k1.PrivateKey) (string, []string, error) {
	depositTxHash := depositTx.TxHash()

	vaultOutIndex := 0
	vaultOut := depositTx.TxOut[vaultOutIndex]

	liquidationCet := wire.NewMsgTx(2)
	liquidationCet.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&depositTxHash, uint32(vaultOutIndex)), nil, nil))
	liquidationCet.AddTxOut(wire.NewTxOut(vaultOut.Value-1000, dcmPkScript))

	p, err := psbt.NewFromUnsignedTx(liquidationCet)
	if err != nil {
		fmt.Printf("failed to create liquidation cet psbt: %v\n", err)
		return "", nil, err
	}

	p.Inputs[0].WitnessUtxo = depositTx.TxOut[vaultOutIndex]
	p.Inputs[0].SighashType = txscript.SigHashDefault

	liquidationCetPsbt, err := p.B64Encode()
	if err != nil {
		fmt.Printf("failed to serialize liquidation cet psbt: %v\n", err)
		return "", nil, err
	}

	fmt.Printf("liquidation cet psbt: %s\n", liquidationCetPsbt)

	script, err := hex.DecodeString(cetInfo.Script.Script)
	if err != nil {
		fmt.Printf("failed to decode script: %v\n", err)
		return "", nil, err
	}

	signaturePoint, err := hex.DecodeString(cetInfo.SignaturePoint)
	if err != nil {
		fmt.Printf("failed to decode signature point: %v\n", err)
		return "", nil, err
	}

	sigHash, err := lendingtypes.CalcTapscriptSigHash(p, 0, txscript.SigHashDefault, script)
	if err != nil {
		fmt.Printf("failed to calculate sig hash: %v\n", err)
		return "", nil, err
	}

	adaptorSig, err := adaptor.Sign(privKey, sigHash, signaturePoint)
	if err != nil {
		fmt.Printf("failed to get the adaptor signature: %v\n", err)
		return "", nil, err
	}

	adaptorSigHex := hex.EncodeToString(adaptorSig.Serialize())

	fmt.Printf("adaptor signature: %s\n", adaptorSigHex)
	fmt.Printf("adaptor signature verified: %t\n", adaptor.Verify(adaptorSig.Serialize(), sigHash, schnorr.SerializePubKey(privKey.PubKey()), signaturePoint))

	return liquidationCetPsbt, []string{adaptorSigHex}, nil
}

func buildRepaymentCet(depositTx *wire.MsgTx, borrowerPkScript []byte, cetInfo *lendingtypes.CetInfo, privKey *secp256k1.PrivateKey) (string, []string, error) {
	depositTxHash := depositTx.TxHash()

	vaultOutIndex := 0
	vaultOut := depositTx.TxOut[vaultOutIndex]

	repaymentCet := wire.NewMsgTx(2)
	repaymentCet.AddTxIn(wire.NewTxIn(wire.NewOutPoint(&depositTxHash, uint32(vaultOutIndex)), nil, nil))
	repaymentCet.AddTxOut(wire.NewTxOut(vaultOut.Value-1000, borrowerPkScript))

	p, err := psbt.NewFromUnsignedTx(repaymentCet)
	if err != nil {
		fmt.Printf("failed to create repayment cet psbt: %v\n", err)
		return "", nil, err
	}

	p.Inputs[0].WitnessUtxo = depositTx.TxOut[vaultOutIndex]
	p.Inputs[0].SighashType = txscript.SigHashDefault

	repaymentCetPsbt, err := p.B64Encode()
	if err != nil {
		fmt.Printf("failed to serialize repayment cet psbt: %v\n", err)
		return "", nil, err
	}

	fmt.Printf("repayment cet psbt: %s\n", repaymentCetPsbt)

	script, err := hex.DecodeString(cetInfo.Script.Script)
	if err != nil {
		fmt.Printf("failed to decode script: %v\n", err)
		return "", nil, err
	}

	sigHash, err := lendingtypes.CalcTapscriptSigHash(p, 0, txscript.SigHashDefault, script)
	if err != nil {
		fmt.Printf("failed to calculate sig hash: %v\n", err)
		return "", nil, err
	}

	schnorrSig, err := schnorr.Sign(privKey, sigHash)
	if err != nil {
		fmt.Printf("failed to get the schnorr signature: %v\n", err)
		return "", nil, err
	}

	schnorrSigHex := hex.EncodeToString(schnorrSig.Serialize())

	fmt.Printf("schnorr signature: %s\n", schnorrSigHex)
	fmt.Printf("schnorr signature verified: %t\n", schnorrSig.Verify(sigHash, privKey.PubKey()))

	return repaymentCetPsbt, []string{schnorrSigHex}, nil
}

func serializeTxB64(tx *wire.MsgTx) string {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())
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
