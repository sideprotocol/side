package types

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/sideprotocol/side/crypto/adaptor"
	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
)

// BuildDLCMeta creates the dlc meta from the given params
func BuildDLCMeta(depositTx *psbt.Packet, vaultPkScript []byte, liquidationCet string, liquidationAdaptorSig string, borrowerPubKey string, agencyPubKey string, secretHash string, muturityTime int64, finalTimeout int64) (*DLCMeta, error) {
	vaultUtxo, err := getVaultUTXO(depositTx, vaultPkScript)
	if err != nil {
		return nil, err
	}

	liquidationCetPacket, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidationCet)), true)
	if err != nil {
		return nil, err
	}

	multiSigScript, err := CreateMultisigScript([]string{borrowerPubKey, agencyPubKey})
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

	merkleTree := GetTapscriptTree([][]byte{
		multiSigScript, forcedRepaymentScript, timeoutRefundScript,
	})

	multiSigScriptProof := merkleTree.LeafMerkleProofs[0]

	internalKey := GetInternalKey()
	controlBlock, err := GetControlBlock(internalKey, multiSigScriptProof)
	if err != nil {
		return nil, err
	}

	for i := range liquidationCetPacket.Inputs {
		liquidationCetPacket.Inputs[i].SighashType = txscript.SigHashDefault
		liquidationCetPacket.Inputs[i].TaprootInternalKey = schnorr.SerializePubKey(internalKey)
		liquidationCetPacket.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: controlBlock,
				Script:       multiSigScript,
				LeafVersion:  txscript.BaseLeafVersion,
			},
		}
	}

	liquidationCet, err = liquidationCetPacket.B64Encode()
	if err != nil {
		return nil, err
	}

	return &DLCMeta{
		LiquidationCet:              liquidationCet,
		LiquidationAdaptorSignature: liquidationAdaptorSig,
		VaultUtxo:                   vaultUtxo,
		InternalKey:                 hex.EncodeToString(internalKey.SerializeCompressed()),
		LiquidationCetScript:        hex.EncodeToString(multiSigScript),
		RepaymentScript:             hex.EncodeToString(multiSigScript),
		ForcedRepaymentScript:       hex.EncodeToString(forcedRepaymentScript),
		TimeoutRefundScript:         hex.EncodeToString(timeoutRefundScript),
	}, nil
}

// VerifyLiquidationCET verifies the given liquidation cet and corresponding adaptor signature
func VerifyLiquidationCET(depositTx *psbt.Packet, liquidationCET string, borrowerPubKey string, agencyPubKey string, adaptorSignature string, adaptorPoint string) error {
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

	if p.Inputs[0].WitnessUtxo == nil {
		return ErrInvalidCET
	}

	multiSigScript, err := CreateMultisigScript([]string{borrowerPubKey, agencyPubKey})
	if err != nil {
		return err
	}

	sigHash, err := CalcTapscriptSigHash(p, 0, DefaultSigHashType, multiSigScript)
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
func CreateLiquidationCET(depositTx *psbt.Packet, vaultPkScript []byte, agencyPkScript []byte, internalKeyBytes []byte, tapscripts [][]byte, feeRate int64) (string, error) {
	vaultUtxo, err := getVaultUTXO(depositTx, vaultPkScript)
	if err != nil {
		return "", err
	}

	p, err := BuildPsbt([]*btcbridgetypes.UTXO{vaultUtxo}, agencyPkScript, feeRate)
	if err != nil {
		return "", err
	}

	internalKey, err := secp256k1.ParsePubKey(internalKeyBytes)
	if err != nil {
		return "", err
	}

	merkleTree := GetTapscriptTree(tapscripts)
	multiSigScriptProof := merkleTree.LeafMerkleProofs[0]

	controlBlock, err := GetControlBlock(internalKey, multiSigScriptProof)
	if err != nil {
		return "", err
	}

	for i := range p.Inputs {
		p.Inputs[i].TaprootInternalKey = schnorr.SerializePubKey(internalKey)
		p.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: controlBlock,
				Script:       tapscripts[0],
				LeafVersion:  txscript.BaseLeafVersion,
			},
		}
	}

	psbtB64, err := p.B64Encode()
	if err != nil {
		return "", err
	}

	return psbtB64, nil
}

// CreateRepaymentTransaction creates the repayment transaction
func CreateRepaymentTransaction(depositTx *psbt.Packet, vaultPkScript []byte, borrowerPkScript []byte, internalKeyBytes []byte, tapscripts [][]byte, feeRate int64) (string, error) {
	vaultUtxo, err := getVaultUTXO(depositTx, vaultPkScript)
	if err != nil {
		return "", err
	}

	p, err := BuildPsbt([]*btcbridgetypes.UTXO{vaultUtxo}, borrowerPkScript, feeRate)
	if err != nil {
		return "", err
	}

	internalKey, err := secp256k1.ParsePubKey(internalKeyBytes)
	if err != nil {
		return "", err
	}

	merkleTree := GetTapscriptTree(tapscripts)
	multiSigScriptProof := merkleTree.LeafMerkleProofs[0]

	controlBlock, err := GetControlBlock(internalKey, multiSigScriptProof)
	if err != nil {
		return "", err
	}

	for i := range p.Inputs {
		p.Inputs[i].TaprootInternalKey = schnorr.SerializePubKey(internalKey)
		p.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: controlBlock,
				Script:       tapscripts[0],
				LeafVersion:  txscript.BaseLeafVersion,
			},
		}
	}

	psbtB64, err := p.B64Encode()
	if err != nil {
		return "", err
	}

	return psbtB64, nil
}

// CreateForcedRepaymentTransaction creates the forced repayment tx
func CreateForcedRepaymentTransaction(depositTx *psbt.Packet, vaultPkScript []byte, agencyPkScript []byte, internalKey []byte, tapscripts [][]byte, feeRate int64) (string, error) {
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

	psbtB64, err := p.B64Encode()
	if err != nil {
		return "", err
	}

	return psbtB64, nil
}

// CreateTimeoutRefundTransaction creates the timeout refund tx
func CreateTimeoutRefundTransaction(depositTx *psbt.Packet, vaultPkScript []byte, borrowerPkScript []byte, internalKey []byte, tapscripts [][]byte, feeRate int64) (string, error) {
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

	psbtB64, err := p.B64Encode()
	if err != nil {
		return "", err
	}

	return psbtB64, nil
}

// BuildSignedLiquidationCet builds the signed liquidation cet from the given signatures
func BuildSignedLiquidationCet(liquidationCet string, borrowerPubKey string, borrowerSignatures []string, agencyPubKey string, agencySignatures []string) ([]byte, error) {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidationCet)), true)
	if err != nil {
		return nil, err
	}

	borrowerPubKeyBytes, err := hex.DecodeString(borrowerPubKey)
	if err != nil {
		return nil, err
	}

	agencyPubKeyBytes, err := hex.DecodeString(agencyPubKey)
	if err != nil {
		return nil, err
	}

	borrowerSig, err := hex.DecodeString(borrowerSignatures[0])
	if err != nil {
		return nil, err
	}

	agencySig, err := hex.DecodeString(agencySignatures[0])
	if err != nil {
		return nil, err
	}

	leafHash := txscript.NewBaseTapLeaf(p.Inputs[0].TaprootLeafScript[0].Script).TapHash()

	p.Inputs[0].TaprootScriptSpendSig = []*psbt.TaprootScriptSpendSig{
		{
			XOnlyPubKey: agencyPubKeyBytes,
			LeafHash:    leafHash[:],
			Signature:   agencySig,
			SigHash:     txscript.SigHashDefault,
		},
		{
			XOnlyPubKey: borrowerPubKeyBytes,
			LeafHash:    leafHash[:],
			Signature:   borrowerSig,
			SigHash:     txscript.SigHashDefault,
		},
	}

	if err := psbt.MaybeFinalizeAll(p); err != nil {
		return nil, err
	}

	signedTx, err := psbt.Extract(p)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := signedTx.Serialize(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetLiquidationCetSigHashes gets the sig hashes of the liquidation cet
func GetLiquidationCetSigHashes(dlcMeta *DLCMeta) ([]string, error) {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.LiquidationCet)), true)
	if err != nil {
		return nil, err
	}

	script, err := hex.DecodeString(dlcMeta.LiquidationCetScript)
	if err != nil {
		return nil, err
	}

	sigHashes := []string{}

	for i, input := range p.Inputs {
		sigHash, err := CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return nil, err
		}

		sigHashes = append(sigHashes, hex.EncodeToString(sigHash))
	}

	return sigHashes, nil
}

// GetDLCTapscripts gets the tap scripts from the given dlc meta
// Assume that the dlc meta is valid
func GetDLCTapscripts(dlcMeta *DLCMeta) [][]byte {
	multiSigScript, _ := hex.DecodeString(dlcMeta.LiquidationCetScript)
	forcedRepaymentScript, _ := hex.DecodeString(dlcMeta.ForcedRepaymentScript)
	timeoutRefundScript, _ := hex.DecodeString(dlcMeta.TimeoutRefundScript)

	return [][]byte{multiSigScript, forcedRepaymentScript, timeoutRefundScript}
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

// getVaultUTXO gets the vault utxo from the given params
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
