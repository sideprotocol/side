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
func BuildDLCMeta(depositTxs []*psbt.Packet, vaultPkScript []byte, liquidationCet string, liquidationAdaptorSignatures []string, defaultLiquidationAdaptorSignatures []string, repaymentCet string, repaymentSignatures []string, borrowerPubKey string, dcmPubKey string, muturityTime int64, finalTimeout int64) (*DLCMeta, error) {
	vaultUtxos, err := getVaultUtxos(depositTxs, vaultPkScript)
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

	multisigScript, err := CreateMultisigScript([]string{borrowerPubKey, dcmPubKey})
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

	borrowerPkScript, err := GetPkScriptFromPubKey(borrowerPubKey)
	if err != nil {
		return nil, err
	}

	timeoutRefundTx, err := CreateTimeoutRefundTransaction(depositTxs, vaultPkScript, borrowerPkScript, internalKey.SerializeCompressed(), [][]byte{multisigScript, timeoutRefundScript}, 1)
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
		TimeoutRefundTx:     timeoutRefundTx,
		VaultUtxos:          vaultUtxos,
		InternalKey:         hex.EncodeToString(internalKey.SerializeCompressed()),
		MultisigScript:      hex.EncodeToString(multisigScript),
		TimeoutRefundScript: hex.EncodeToString(timeoutRefundScript),
	}, nil
}

// VerifyCets verifies the given cets
func VerifyCets(depositTxs []*psbt.Packet, vaultPkScript []byte, borrowerPubKey string, dcmPubKey string, liquidationEvent *dlctypes.DLCEvent, defaultLiquidationEvent *dlctypes.DLCEvent, liquidationCet string, liquidationAdaptorSignatures []string, defaultLiquidationAdaptorSignatures []string, repaymentCet string, repaymentSignatures []string) error {
	liquidationAdaptorPoint, err := dlctypes.GetSignaturePointFromEvent(liquidationEvent, 0)
	if err != nil {
		return err
	}

	defaultLiquidationAdaptorPoint, err := dlctypes.GetSignaturePointFromEvent(defaultLiquidationEvent, 0)
	if err != nil {
		return err
	}

	if err := VerifyLiquidationCet(depositTxs, vaultPkScript, borrowerPubKey, dcmPubKey, liquidationCet, liquidationAdaptorSignatures, liquidationAdaptorPoint); err != nil {
		return err
	}

	if err := VerifyLiquidationCet(depositTxs, vaultPkScript, borrowerPubKey, dcmPubKey, liquidationCet, defaultLiquidationAdaptorSignatures, defaultLiquidationAdaptorPoint); err != nil {
		return err
	}

	if err := VerifyRepaymentCet(depositTxs, vaultPkScript, borrowerPubKey, dcmPubKey, repaymentCet, repaymentSignatures); err != nil {
		return err
	}

	return nil
}

// VerifyLiquidationCet verifies the given liquidation cet and corresponding adaptor signatures
func VerifyLiquidationCet(depositTxs []*psbt.Packet, vaultPkScript []byte, borrowerPubKey string, dcmPubKey string, liquidationCET string, adaptorSignatures []string, adaptorPoint []byte) error {
	vaultUtxos, err := getVaultUtxos(depositTxs, vaultPkScript)
	if err != nil {
		return err
	}

	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidationCET)), true)
	if err != nil {
		return ErrInvalidCET
	}

	if len(p.UnsignedTx.TxIn) != len(vaultUtxos) {
		return errorsmod.Wrap(ErrInvalidCET, "incorrect input number")
	}

	for i, txIn := range p.UnsignedTx.TxIn {
		if txIn.PreviousOutPoint.Hash.String() != vaultUtxos[i].Txid {
			return errorsmod.Wrap(ErrInvalidCET, "incorrect previous tx hash")
		}

		if txIn.PreviousOutPoint.Index != uint32(vaultUtxos[i].Vout) {
			return errorsmod.Wrap(ErrInvalidCET, "incorrect previous tx out index")
		}

		if p.Inputs[i].WitnessUtxo == nil {
			return errorsmod.Wrap(ErrInvalidCET, "missing witness utxo")
		}

		if !bytes.Equal(p.Inputs[i].WitnessUtxo.PkScript, vaultPkScript) {
			return errorsmod.Wrap(ErrInvalidCET, "mismatched witness utxo pk script")
		}

		if p.Inputs[i].WitnessUtxo.Value != int64(vaultUtxos[i].Amount) {
			return errorsmod.Wrap(ErrInvalidCET, "mismatched witness utxo value")
		}
	}

	if len(adaptorSignatures) != len(p.Inputs) {
		return errorsmod.Wrap(ErrInvalidAdaptorSignatures, "incorrect signature number")
	}

	pubKeyBytes, err := hex.DecodeString(borrowerPubKey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode borrower public key")
	}

	script, err := CreateMultisigScript([]string{borrowerPubKey, dcmPubKey})
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
func VerifyRepaymentCet(depositTxs []*psbt.Packet, vaultPkScript []byte, borrowerPubKey string, dcmPubKey string, repaymentCet string, signatures []string) error {
	vaultUtxos, err := getVaultUtxos(depositTxs, vaultPkScript)
	if err != nil {
		return err
	}

	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(repaymentCet)), true)
	if err != nil {
		return ErrInvalidCET
	}

	if len(p.UnsignedTx.TxIn) != len(vaultUtxos) {
		return errorsmod.Wrap(ErrInvalidCET, "incorrect input number")
	}

	for i, txIn := range p.UnsignedTx.TxIn {
		if txIn.PreviousOutPoint.Hash.String() != vaultUtxos[i].Txid {
			return errorsmod.Wrap(ErrInvalidCET, "incorrect previous tx hash")
		}

		if txIn.PreviousOutPoint.Index != uint32(vaultUtxos[i].Vout) {
			return errorsmod.Wrap(ErrInvalidCET, "incorrect previous tx out index")
		}

		if p.Inputs[i].WitnessUtxo == nil {
			return errorsmod.Wrap(ErrInvalidCET, "missing witness utxo")
		}

		if !bytes.Equal(p.Inputs[i].WitnessUtxo.PkScript, vaultPkScript) {
			return errorsmod.Wrap(ErrInvalidCET, "mismatched witness utxo pk script")
		}

		if p.Inputs[i].WitnessUtxo.Value != int64(vaultUtxos[i].Amount) {
			return errorsmod.Wrap(ErrInvalidCET, "mismatched witness utxo value")
		}
	}

	if len(signatures) != len(p.Inputs) {
		return errorsmod.Wrap(ErrInvalidSignatures, "incorrect signature number")
	}

	pubKeyBytes, err := hex.DecodeString(borrowerPubKey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode borrower public key")
	}

	script, err := CreateMultisigScript([]string{borrowerPubKey, dcmPubKey})
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
func CreateLiquidationCET(depositTxs []*psbt.Packet, vaultPkScript []byte, dcmPkScript []byte, internalKeyBytes []byte, tapscripts [][]byte, feeRate int64) (string, error) {
	vaultUtxos, err := getVaultUtxos(depositTxs, vaultPkScript)
	if err != nil {
		return "", err
	}

	p, err := BuildPsbt(vaultUtxos, dcmPkScript, feeRate)
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
func CreateRepaymentCet(depositTxs []*psbt.Packet, vaultPkScript []byte, borrowerPkScript []byte, internalKeyBytes []byte, tapscripts [][]byte, feeRate int64) (string, error) {
	vaultUtxos, err := getVaultUtxos(depositTxs, vaultPkScript)
	if err != nil {
		return "", err
	}

	p, err := BuildPsbt(vaultUtxos, borrowerPkScript, feeRate)
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
func CreateDefaultLiquidationCet(depositTxs []*psbt.Packet, vaultPkScript []byte, dcmPkScript []byte, internalKeyBytes []byte, tapscripts [][]byte, feeRate int64) (string, error) {
	vaultUtxos, err := getVaultUtxos(depositTxs, vaultPkScript)
	if err != nil {
		return "", err
	}

	p, err := BuildPsbt(vaultUtxos, dcmPkScript, feeRate)
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

// CreateTimeoutRefundTransaction creates the timeout refund tx
func CreateTimeoutRefundTransaction(depositTxs []*psbt.Packet, vaultPkScript []byte, borrowerPkScript []byte, internalKeyBytes []byte, tapscripts [][]byte, feeRate int64) (string, error) {
	vaultUtxos, err := getVaultUtxos(depositTxs, vaultPkScript)
	if err != nil {
		return "", err
	}

	p, err := BuildPsbt(vaultUtxos, borrowerPkScript, feeRate)
	if err != nil {
		return "", err
	}

	internalKey, err := secp256k1.ParsePubKey(internalKeyBytes)
	if err != nil {
		return "", err
	}

	merkleTree := GetTapscriptTree(tapscripts)
	timeoutRefundScriptProof := merkleTree.LeafMerkleProofs[1]

	controlBlock, err := GetControlBlock(internalKey, timeoutRefundScriptProof)
	if err != nil {
		return "", err
	}

	for i := range p.Inputs {
		p.Inputs[i].TaprootInternalKey = btcschnorr.SerializePubKey(internalKey)
		p.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: controlBlock,
				Script:       tapscripts[1],
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

// BuildSignedCet builds the signed cet from the given signatures
// Assume that the cet is valid and signatures match
func BuildSignedCet(cet string, borrowerPubKey string, borrowerSignatures []string, dcmPubKey string, dcmSignatures []string) ([]byte, *chainhash.Hash, error) {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(cet)), true)
	if err != nil {
		return nil, nil, err
	}

	borrowerPubKeyBytes, err := hex.DecodeString(borrowerPubKey)
	if err != nil {
		return nil, nil, err
	}

	dcmPubKeyBytes, err := hex.DecodeString(dcmPubKey)
	if err != nil {
		return nil, nil, err
	}

	for i, input := range p.Inputs {
		borrowerSig, err := hex.DecodeString(borrowerSignatures[i])
		if err != nil {
			return nil, nil, err
		}

		dcmSig, err := hex.DecodeString(dcmSignatures[i])
		if err != nil {
			return nil, nil, err
		}

		leafHash := txscript.NewBaseTapLeaf(input.TaprootLeafScript[0].Script).TapHash()

		p.Inputs[i].TaprootScriptSpendSig = []*psbt.TaprootScriptSpendSig{
			{
				XOnlyPubKey: dcmPubKeyBytes,
				LeafHash:    leafHash[:],
				Signature:   dcmSig,
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

// GetDefaultLiquidationCetSigHashes gets the sig hashes of the default liquidation cet
func GetDefaultLiquidationCetSigHashes(dlcMeta *DLCMeta) ([]string, error) {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.DefaultLiquidationCet.Tx)), true)
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

// GetLiquidationCetOutput gets the output value for the given liquidation cet
// Assume that the given cet is valid
func GetLiquidationCetOutput(liquidationCet string) int64 {
	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidationCet)), true)

	return p.UnsignedTx.TxOut[0].Value
}

// GetDLCTapscripts gets the tap scripts from the given dlc meta
// Assume that the dlc meta is valid
func GetDLCTapscripts(dlcMeta *DLCMeta) [][]byte {
	multisigScript, _ := hex.DecodeString(dlcMeta.MultisigScript)
	timeoutRefundScript, _ := hex.DecodeString(dlcMeta.TimeoutRefundScript)

	return [][]byte{multisigScript, timeoutRefundScript}
}

// getVaultUtxosFromDepositTx gets vault utxos from the given deposit tx
func getVaultUtxosFromDepositTx(depositTx *psbt.Packet, vaultPkScript []byte) ([]*btcbridgetypes.UTXO, error) {
	utxos := []*btcbridgetypes.UTXO{}

	found := false

	for i, out := range depositTx.UnsignedTx.TxOut {
		if bytes.Equal(out.PkScript, vaultPkScript) {
			utxo := &btcbridgetypes.UTXO{
				Txid:         depositTx.UnsignedTx.TxHash().String(),
				Vout:         uint64(i),
				Amount:       uint64(out.Value),
				PubKeyScript: out.PkScript,
			}

			utxos = append(utxos, utxo)

			found = true
		}
	}

	if !found {
		return nil, ErrInvalidDepositTx
	}

	return utxos, nil
}

// getVaultUtxos gets the vault utxos from the given deposit txs
func getVaultUtxos(depositTxs []*psbt.Packet, vaultPkScript []byte) ([]*btcbridgetypes.UTXO, error) {
	utxos := []*btcbridgetypes.UTXO{}

	for _, depositTx := range depositTxs {
		vaultUtxos, err := getVaultUtxosFromDepositTx(depositTx, vaultPkScript)
		if err != nil {
			return nil, err
		}

		utxos = append(utxos, vaultUtxos...)
	}

	return utxos, nil
}
