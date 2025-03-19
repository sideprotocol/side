package types

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"

	btcschnorr "github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	errorsmod "cosmossdk.io/errors"

	"github.com/sideprotocol/side/crypto/adaptor"
	"github.com/sideprotocol/side/crypto/schnorr"
	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
)

// BuildDLCMeta creates the dlc meta from the given params
func BuildDLCMeta(depositTx *psbt.Packet, vaultPkScript []byte, liquidationCet string, liquidationAdaptorSignatures []string, defaultLiquidationAdaptorSignatures []string, repaymentCet string, repaymentSignatures []string, borrowerPubKey string, agencyPubKey string, muturityTime int64, finalTimeout int64) (*DLCMeta, error) {
	vaultUtxo, err := getVaultUTXO(depositTx, vaultPkScript)
	if err != nil {
		return nil, err
	}

	liquidationCetPsbt, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidationCet)), true)
	if err != nil {
		return nil, err
	}

	repaymentCetPsbt, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(repaymentCet)), true)
	if err != nil {
		return nil, err
	}

	multisigScript, err := CreateMultisigScript([]string{borrowerPubKey, agencyPubKey})
	if err != nil {
		return nil, err
	}

	timeoutRefundScript, err := CreatePubKeyTimeLockScript(borrowerPubKey, finalTimeout)
	if err != nil {
		return nil, err
	}

	merkleTree := GetTapscriptTree([][]byte{
		multisigScript, timeoutRefundScript,
	})

	multisigScriptProof := merkleTree.LeafMerkleProofs[0]

	internalKey := GetInternalKey()
	controlBlock, err := GetControlBlock(internalKey, multisigScriptProof)
	if err != nil {
		return nil, err
	}

	for i := range liquidationCetPsbt.Inputs {
		liquidationCetPsbt.Inputs[i].SighashType = txscript.SigHashDefault
		liquidationCetPsbt.Inputs[i].TaprootInternalKey = btcschnorr.SerializePubKey(internalKey)
		liquidationCetPsbt.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: controlBlock,
				Script:       multisigScript,
				LeafVersion:  txscript.BaseLeafVersion,
			},
		}
	}

	for i := range repaymentCetPsbt.Inputs {
		repaymentCetPsbt.Inputs[i].SighashType = txscript.SigHashDefault
		repaymentCetPsbt.Inputs[i].TaprootInternalKey = btcschnorr.SerializePubKey(internalKey)
		repaymentCetPsbt.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: controlBlock,
				Script:       multisigScript,
				LeafVersion:  txscript.BaseLeafVersion,
			},
		}
	}

	liquidationCet, err = liquidationCetPsbt.B64Encode()
	if err != nil {
		return nil, err
	}

	repaymentCet, err = repaymentCetPsbt.B64Encode()
	if err != nil {
		return nil, err
	}

	return &DLCMeta{
		LiquidationCet: LiquidationCet{
			Tx:                        liquidationCet,
			BorrowerAdaptorSignatures: liquidationAdaptorSignatures,
		},
		DefaultLiquidationCet: LiquidationCet{
			Tx:                        liquidationCet,
			BorrowerAdaptorSignatures: defaultLiquidationAdaptorSignatures,
		},
		RepaymentCet: RepaymentCet{
			Tx:                 repaymentCet,
			BorrowerSignatures: repaymentSignatures,
		},
		VaultUtxos:          []*btcbridgetypes.UTXO{vaultUtxo},
		InternalKey:         hex.EncodeToString(internalKey.SerializeCompressed()),
		MultisigScript:      hex.EncodeToString(multisigScript),
		TimeoutRefundScript: hex.EncodeToString(timeoutRefundScript),
	}, nil
}

// VerifyCets verifies the given cets
func VerifyCets(depositTx *psbt.Packet, borrowerPubKey string, agencyPubKey string, liquidationEvent *dlctypes.DLCEvent, defaultLiquidationEvent *dlctypes.DLCEvent, liquidationCet string, liquidationAdaptorSignatures []string, defaultLiquidationAdaptorSignatures []string, repaymentCet string, repaymentSignatures []string) error {
	liquidationAdaptorPoint, err := dlctypes.GetSignaturePointFromEvent(liquidationEvent, 0)
	if err != nil {
		return err
	}

	defaultLiquidationAdaptorPoint, err := dlctypes.GetSignaturePointFromEvent(defaultLiquidationEvent, 0)
	if err != nil {
		return err
	}

	if err := VerifyLiquidationCet(depositTx, borrowerPubKey, agencyPubKey, liquidationCet, liquidationAdaptorSignatures, liquidationAdaptorPoint); err != nil {
		return err
	}

	if err := VerifyLiquidationCet(depositTx, borrowerPubKey, agencyPubKey, liquidationCet, defaultLiquidationAdaptorSignatures, defaultLiquidationAdaptorPoint); err != nil {
		return err
	}

	if err := VerifyRepaymentCet(depositTx, borrowerPubKey, agencyPubKey, repaymentCet, repaymentSignatures); err != nil {
		return err
	}

	return nil
}

// VerifyLiquidationCet verifies the given liquidation cet and corresponding adaptor signatures
func VerifyLiquidationCet(depositTx *psbt.Packet, borrowerPubKey string, agencyPubKey string, liquidationCET string, adaptorSignatures []string, adaptorPoint []byte) error {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidationCET)), true)
	if err != nil {
		return ErrInvalidCET
	}

	depositTxHash := depositTx.UnsignedTx.TxHash()

	for i, input := range p.UnsignedTx.TxIn {
		if !input.PreviousOutPoint.Hash.IsEqual(&depositTxHash) {
			return errorsmod.Wrap(ErrInvalidCET, "incorrect previous tx hash")
		}

		if p.Inputs[i].WitnessUtxo == nil {
			return errorsmod.Wrap(ErrInvalidCET, "missing witness utxo")
		}
	}

	if len(adaptorSignatures) != len(p.Inputs) {
		return errorsmod.Wrap(ErrInvalidAdaptorSignatures, "incorrect signature number")
	}

	pubKeyBytes, err := hex.DecodeString(borrowerPubKey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode borrower public key")
	}

	script, err := CreateMultisigScript([]string{borrowerPubKey, agencyPubKey})
	if err != nil {
		return err
	}

	for i, signature := range adaptorSignatures {
		sigHash, err := CalcTapscriptSigHash(p, i, DefaultSigHashType, script)
		if err != nil {
			return errorsmod.Wrapf(err, "failed to calculate sig hash")
		}

		sigBytes, err := hex.DecodeString(signature)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidAdaptorSignature, "failed to decode adaptor signature")
		}

		if !adaptor.Verify(sigBytes, sigHash, pubKeyBytes, adaptorPoint) {
			return ErrInvalidAdaptorSignature
		}
	}

	return nil
}

// VerifyRepaymentCet verifies the given repayment cet and corresponding signatures
func VerifyRepaymentCet(depositTx *psbt.Packet, borrowerPubKey string, agencyPubKey string, repaymentCet string, signatures []string) error {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(repaymentCet)), true)
	if err != nil {
		return ErrInvalidCET
	}

	depositTxHash := depositTx.UnsignedTx.TxHash()

	for i, input := range p.UnsignedTx.TxIn {
		if !input.PreviousOutPoint.Hash.IsEqual(&depositTxHash) {
			return errorsmod.Wrap(ErrInvalidCET, "incorrect previous tx hash")
		}

		if p.Inputs[i].WitnessUtxo == nil {
			return errorsmod.Wrap(ErrInvalidCET, "missing witness utxo")
		}
	}

	if len(signatures) != len(p.Inputs) {
		return errorsmod.Wrap(ErrInvalidSignatures, "incorrect signature number")
	}

	pubKeyBytes, err := hex.DecodeString(borrowerPubKey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode borrower public key")
	}

	script, err := CreateMultisigScript([]string{borrowerPubKey, agencyPubKey})
	if err != nil {
		return err
	}

	for i, signature := range signatures {
		sigHash, err := CalcTapscriptSigHash(p, i, DefaultSigHashType, script)
		if err != nil {
			return errorsmod.Wrapf(err, "failed to calculate sig hash")
		}

		sigBytes, err := hex.DecodeString(signature)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidSignature, "failed to decode adaptor signature")
		}

		if !schnorr.Verify(sigBytes, sigHash, pubKeyBytes) {
			return ErrInvalidSignature
		}
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
		p.Inputs[i].TaprootInternalKey = btcschnorr.SerializePubKey(internalKey)
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

// CreateRepaymentCet creates the repayment cet
func CreateRepaymentCet(depositTx *psbt.Packet, vaultPkScript []byte, borrowerPkScript []byte, internalKeyBytes []byte, tapscripts [][]byte, feeRate int64) (string, error) {
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
		p.Inputs[i].TaprootInternalKey = btcschnorr.SerializePubKey(internalKey)
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

// CreateDefaultLiquidationCet creates the default liquidation cet
func CreateDefaultLiquidationCet(depositTx *psbt.Packet, vaultPkScript []byte, agencyPkScript []byte, internalKey []byte, tapscripts [][]byte, feeRate int64) (string, error) {
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

// BuildSignedCet builds the signed cet from the given signatures
// Assume that the cet is valid and signatures match
func BuildSignedCet(cet string, borrowerPubKey string, borrowerSignatures []string, agencyPubKey string, agencySignatures []string) ([]byte, *chainhash.Hash, error) {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(cet)), true)
	if err != nil {
		return nil, nil, err
	}

	borrowerPubKeyBytes, err := hex.DecodeString(borrowerPubKey)
	if err != nil {
		return nil, nil, err
	}

	agencyPubKeyBytes, err := hex.DecodeString(agencyPubKey)
	if err != nil {
		return nil, nil, err
	}

	for i, input := range p.Inputs {
		borrowerSig, err := hex.DecodeString(borrowerSignatures[i])
		if err != nil {
			return nil, nil, err
		}

		agencySig, err := hex.DecodeString(agencySignatures[i])
		if err != nil {
			return nil, nil, err
		}

		leafHash := txscript.NewBaseTapLeaf(input.TaprootLeafScript[0].Script).TapHash()

		p.Inputs[i].TaprootScriptSpendSig = []*psbt.TaprootScriptSpendSig{
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
	}

	if err := psbt.MaybeFinalizeAll(p); err != nil {
		return nil, nil, err
	}

	signedTx, err := psbt.Extract(p)
	if err != nil {
		return nil, nil, err
	}

	var buf bytes.Buffer
	if err := signedTx.Serialize(&buf); err != nil {
		return nil, nil, err
	}

	txHash := signedTx.TxHash()

	return buf.Bytes(), &txHash, nil
}

// GetCetInfo gets the cet info from the given event and script
func GetCetInfo(event *dlctypes.DLCEvent, outcomeIndex int, script []byte) (*CetInfo, error) {
	if event == nil {
		return nil, nil
	}

	signaturePoint, err := dlctypes.GetSignaturePointFromEvent(event, outcomeIndex)
	if err != nil {
		return nil, err
	}

	return &CetInfo{
		EventId:        event.Id,
		OutcomeIndex:   uint32(outcomeIndex),
		SignaturePoint: hex.EncodeToString(signaturePoint),
		Script:         hex.EncodeToString(script),
	}, nil
}

// GetLiquidationCetSigHashes gets the sig hashes of the liquidation cet
func GetLiquidationCetSigHashes(dlcMeta *DLCMeta) ([]string, error) {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.LiquidationCet.Tx)), true)
	if err != nil {
		return nil, err
	}

	script, err := hex.DecodeString(dlcMeta.MultisigScript)
	if err != nil {
		return nil, err
	}

	sigHashes := []string{}

	for i, input := range p.Inputs {
		sigHash, err := CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return nil, err
		}

		sigHashes = append(sigHashes, base64.StdEncoding.EncodeToString(sigHash))
	}

	return sigHashes, nil
}

// GetRepaymentCetSigHashes gets the sig hashes of the repayment cet
func GetRepaymentCetSigHashes(dlcMeta *DLCMeta) ([]string, error) {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.RepaymentCet.Tx)), true)
	if err != nil {
		return nil, err
	}

	script, err := hex.DecodeString(dlcMeta.MultisigScript)
	if err != nil {
		return nil, err
	}

	sigHashes := []string{}

	for i, input := range p.Inputs {
		sigHash, err := CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return nil, err
		}

		sigHashes = append(sigHashes, base64.StdEncoding.EncodeToString(sigHash))
	}

	return sigHashes, nil
}

// GetDLCTapscripts gets the tap scripts from the given dlc meta
// Assume that the dlc meta is valid
func GetDLCTapscripts(dlcMeta *DLCMeta) [][]byte {
	multisigScript, _ := hex.DecodeString(dlcMeta.MultisigScript)
	timeoutRefundScript, _ := hex.DecodeString(dlcMeta.TimeoutRefundScript)

	return [][]byte{multisigScript, timeoutRefundScript}
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
