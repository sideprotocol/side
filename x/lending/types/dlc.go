package types

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"

	"github.com/sideprotocol/side/crypto/adaptor"
	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
)

// BuildDLCMeta creates the dlc meta from the given params
func BuildDLCMeta(depositTx *psbt.Packet, vaultPkScript []byte, liquidationCET string, liquidationAdaptorSig string, borrowerPubKey string, agencyPubKey string, secretHash string, muturityTime int64, finalTimeout int64) (*DLCMeta, error) {
	vaultUtxo, err := getVaultUTXO(depositTx, vaultPkScript)
	if err != nil {
		return nil, err
	}

	liquidationCETScript, err := CreateMultisigScript([]string{borrowerPubKey, agencyPubKey})
	if err != nil {
		return nil, err
	}

	forcedRepaymentScript, err := CreateHashTimeLockScript(agencyPubKey, secretHash, muturityTime)
	if err != nil {
		return nil, err
	}

	timeoutRefundScript, err := CreatePubKeyTimeLockScript(borrowerPubKey, finalTimeout)
	if err != nil {
		return nil, err
	}

	return &DLCMeta{
		LiquidationCet:              liquidationCET,
		LiquidationAdaptorSignature: liquidationAdaptorSig,
		VaultUtxo:                   vaultUtxo,
		InternalKey:                 hex.EncodeToString(schnorr.SerializePubKey(GetInternalKey())),
		LiquidationCetScript:        hex.EncodeToString(liquidationCETScript),
		ForcedRepaymentScript:       hex.EncodeToString(forcedRepaymentScript),
		TimeoutRefundScript:         hex.EncodeToString(timeoutRefundScript),
		TapscriptMerkleRoot:         GetTapscriptsMerkleRoot([][]byte{liquidationCETScript, forcedRepaymentScript, timeoutRefundScript}),
	}, nil
}

// VerifyLiquidationCET verifies the given liquidation cet and corresponding adaptor signature
func VerifyLiquidationCET(depositTx *psbt.Packet, liquidationCET string, borrowerPubKey string, adaptorSignature string, adaptorPoint string) error {
	if err := depositTx.SanityCheck(); err != nil {
		return ErrInvalidFunding
	}

	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidationCET)), true)
	if err != nil {
		return ErrInvalidCET
	}

	depositTxHash := depositTx.UnsignedTx.TxHash()

	for _, input := range p.UnsignedTx.TxIn {
		if input.PreviousOutPoint.Hash != depositTxHash {
			return ErrInvalidCET
		}
	}

	sigHash, err := txscript.CalcTapscriptSignaturehash(txscript.NewTxSigHashes(p.UnsignedTx, nil), DefaultSigHashType, p.UnsignedTx, 0, nil, txscript.NewBaseTapLeaf(p.Inputs[0].TaprootLeafScript[0].Script), nil)
	if err != nil {
		return ErrInvalidCET
	}

	sigBytes, err := hex.DecodeString(adaptorSignature)
	if err != nil {
		return ErrInvalidAdaptorSignature
	}

	pubKeyBytes, err := hex.DecodeString(borrowerPubKey)
	if err != nil {
		return ErrInvalidBorrowerPubkey
	}

	adaptorPointBytes, err := hex.DecodeString(adaptorPoint)
	if err != nil {
		return ErrInvalidAdaptorPoint
	}

	if !adaptor.Verify(sigBytes, sigHash, pubKeyBytes, adaptorPointBytes) {
		return ErrInvalidAdaptorSignature
	}

	return nil
}

// CreateLiquidationCET creates the liquidation cet
func CreateLiquidationCET(depositTx *psbt.Packet, vaultPkScript []byte, agencyPkScript []byte, internalKey []byte, tapscripts [][]byte, merkleRoot []byte, feeRate int64) (string, error) {
	vaultUtxo, err := getVaultUTXO(depositTx, vaultPkScript)
	if err != nil {
		return "", err
	}

	p, err := BuildPsbt([]*btcbridgetypes.UTXO{vaultUtxo}, agencyPkScript, feeRate)
	if err != nil {
		return "", err
	}

	p.Inputs[0].TaprootInternalKey = internalKey
	p.Inputs[0].TaprootLeafScript = []*psbt.TaprootTapLeafScript{}
	p.Inputs[0].TaprootMerkleRoot = merkleRoot

	psbtB64, err := p.B64Encode()
	if err != nil {
		return "", err
	}

	return psbtB64, nil
}

// CreateForcedRepaymentTransaction creates the forced repayment tx
func CreateForcedRepaymentTransaction(depositTx *psbt.Packet, vaultPkScript []byte, agencyPkScript []byte, internalKey []byte, tapscripts [][]byte, merkleRoot []byte, feeRate int64) (string, error) {
	vaultUtxo, err := getVaultUTXO(depositTx, vaultPkScript)
	if err != nil {
		return "", err
	}

	p, err := BuildPsbt([]*btcbridgetypes.UTXO{vaultUtxo}, agencyPkScript, feeRate)
	if err != nil {
		return "", err
	}

	p.Inputs[0].TaprootInternalKey = internalKey
	p.Inputs[0].TaprootLeafScript = []*psbt.TaprootTapLeafScript{}
	p.Inputs[0].TaprootMerkleRoot = merkleRoot

	psbtB64, err := p.B64Encode()
	if err != nil {
		return "", err
	}

	return psbtB64, nil
}

// CreateTimeoutRefundTransaction creates the timeout refund tx
func CreateTimeoutRefundTransaction(depositTx *psbt.Packet, vaultPkScript []byte, borrowerPkScript []byte, internalKey []byte, tapscripts [][]byte, merkleRoot []byte, feeRate int64) (string, error) {
	vaultUtxo, err := getVaultUTXO(depositTx, vaultPkScript)
	if err != nil {
		return "", err
	}

	p, err := BuildPsbt([]*btcbridgetypes.UTXO{vaultUtxo}, borrowerPkScript, feeRate)
	if err != nil {
		return "", err
	}

	p.Inputs[0].TaprootInternalKey = internalKey
	p.Inputs[0].TaprootLeafScript = []*psbt.TaprootTapLeafScript{}
	p.Inputs[0].TaprootMerkleRoot = merkleRoot

	psbtB64, err := p.B64Encode()
	if err != nil {
		return "", err
	}

	return psbtB64, nil
}

// getVaultOutIndex returns the index of the vault output
func getVaultOutIndex(depositTx *psbt.Packet, vaultPkScript []byte) (int, error) {
	for i, out := range depositTx.UnsignedTx.TxOut {
		if bytes.Equal(out.PkScript, vaultPkScript) {
			return i, nil
		}
	}

	return 0, ErrInvalidDepositTx
}

func getVaultUTXO(depositTx *psbt.Packet, vaultPkScript []byte) (*btcbridgetypes.UTXO, error) {
	vaultOutIndex, err := getVaultOutIndex(depositTx, vaultPkScript)
	if err != nil {
		return nil, err
	}

	vaultOutput := depositTx.UnsignedTx.TxOut[vaultOutIndex]

	return &btcbridgetypes.UTXO{
		Txid:         depositTx.UnsignedTx.TxHash().String(),
		Vout:         uint64(vaultOutIndex),
		Amount:       uint64(vaultOutput.Value),
		PubKeyScript: vaultOutput.PkScript,
	}, nil
}
