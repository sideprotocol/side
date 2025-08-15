package types

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"slices"

	btcschnorr "github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	"github.com/sideprotocol/side/bitcoin/crypto/adaptor"
	"github.com/sideprotocol/side/bitcoin/crypto/schnorr"
	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
)

const (
	// default sig hash type
	DefaultSigHashType = txscript.SigHashDefault

	// liquidation cet sequence
	LiquidationCetSequence = wire.MaxTxInSequenceNum
)

const (
	// outcome index for liquidated
	LiquidatedOutcomeIndex = 0

	// outcome index for default liquidated
	DefaultLiquidatedOutcomeIndex = 1

	// outcome index for repaid
	RepaidOutcomeIndex = 2
)

// BuildDLCMeta builds the dlc meta from the given params
// Assume that the given params are valid
func BuildDLCMeta(borrowerPubKey string, borrowerAuthPubKey string, dcmPubKey string, finalTimeout int64) (*DLCMeta, error) {
	borrowerPubKeyBytes, _ := hex.DecodeString(borrowerPubKey)
	borrowerAuthPubKeyBytes, _ := hex.DecodeString(borrowerAuthPubKey)
	dcmPubKeyBytes, _ := hex.DecodeString(dcmPubKey)

	internalKey := GetInternalKey(borrowerPubKeyBytes, dcmPubKeyBytes)

	liquidationScript, repaymentScript, timeoutRefundScript, _ := GetVaultScripts(borrowerPubKeyBytes, borrowerAuthPubKeyBytes, dcmPubKeyBytes, finalTimeout)

	tapscriptTree := GetTapscriptTree([][]byte{liquidationScript, repaymentScript, timeoutRefundScript})

	liquidationScriptControlBlock, err := GetControlBlock(tapscriptTree, liquidationScript, internalKey)
	if err != nil {
		return nil, err
	}

	repaymentScriptControlBlock, err := GetControlBlock(tapscriptTree, repaymentScript, internalKey)
	if err != nil {
		return nil, err
	}

	timeoutRefundScriptControlBlock, err := GetControlBlock(tapscriptTree, timeoutRefundScript, internalKey)
	if err != nil {
		return nil, err
	}

	return &DLCMeta{
		InternalKey:         hex.EncodeToString(btcschnorr.SerializePubKey(internalKey)),
		LiquidationScript:   GetLeafScript(liquidationScript, liquidationScriptControlBlock),
		RepaymentScript:     GetLeafScript(repaymentScript, repaymentScriptControlBlock),
		TimeoutRefundScript: GetLeafScript(timeoutRefundScript, timeoutRefundScriptControlBlock),
	}, nil
}

// VerifyCets verifies the given cets
func VerifyCets(dlcMeta *DLCMeta, depositTxs []*psbt.Packet, vaultPkScript []byte, borrowerPubKey string, borrowerAuthPubKey string, dcmPubKey string, dlcEvent *dlctypes.DLCEvent, liquidationCet string, liquidationAdaptorSignatures []string, defaultLiquidationAdaptorSignatures []string, repaymentCet string, repaymentSignatures []string, currentFeeRate int64, maxLiquidationFeeRateMultiplier int64) error {
	liquidationAdaptorPoint, err := dlctypes.GetSignaturePointFromEvent(dlcEvent, LiquidatedOutcomeIndex)
	if err != nil {
		return err
	}

	defaultLiquidationAdaptorPoint, err := dlctypes.GetSignaturePointFromEvent(dlcEvent, DefaultLiquidatedOutcomeIndex)
	if err != nil {
		return err
	}

	if err := VerifyLiquidationCet(dlcMeta, depositTxs, vaultPkScript, borrowerAuthPubKey, dcmPubKey, liquidationCet, liquidationAdaptorSignatures, liquidationAdaptorPoint, currentFeeRate, maxLiquidationFeeRateMultiplier); err != nil {
		return errorsmod.Wrapf(ErrInvalidCET, "invalid liquidation cet: %v", err)
	}

	if err := VerifyLiquidationCet(dlcMeta, depositTxs, vaultPkScript, borrowerAuthPubKey, dcmPubKey, liquidationCet, defaultLiquidationAdaptorSignatures, defaultLiquidationAdaptorPoint, currentFeeRate, maxLiquidationFeeRateMultiplier); err != nil {
		return errorsmod.Wrapf(ErrInvalidCET, "invalid default liquidation cet: %v", err)
	}

	if err := VerifyRepaymentCet(dlcMeta, depositTxs, vaultPkScript, borrowerPubKey, dcmPubKey, repaymentCet, repaymentSignatures); err != nil {
		return errorsmod.Wrapf(ErrInvalidCET, "invalid repayment cet: %v", err)
	}

	return nil
}

// VerifyLiquidationCet verifies the given liquidation cet and corresponding adaptor signatures
func VerifyLiquidationCet(dlcMeta *DLCMeta, depositTxs []*psbt.Packet, vaultPkScript []byte, borrowerAuthPubKey string, dcmPubKey string, liquidationCET string, adaptorSignatures []string, adaptorPoint []byte, currentFeeRate int64, maxFeeRateMultiplier int64) error {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidationCET)), true)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidCET, "failed to deserialize cet")
	}

	dcmPkScript, err := GetPkScriptFromPubKey(dcmPubKey)
	if err != nil {
		return err
	}

	if len(p.UnsignedTx.TxOut) != 1 || !bytes.Equal(p.UnsignedTx.TxOut[0].PkScript, dcmPkScript) {
		return errorsmod.Wrap(ErrInvalidCET, "incorrect tx out")
	}

	if btcbridgetypes.IsDustOut(p.UnsignedTx.TxOut[0]) {
		return errorsmod.Wrap(ErrInvalidCET, "dust tx out")
	}

	script, controlBlock, err := UnwrapLeafScript(dlcMeta.LiquidationScript)
	if err != nil {
		return err
	}

	witnessSize := GetCetWitnessSize(CetType_LIQUIDATION, script, controlBlock)

	if err := CheckCetFeeRate(p, witnessSize, currentFeeRate, maxFeeRateMultiplier); err != nil {
		return err
	}

	if err := CheckTransactionWeight(p.UnsignedTx, witnessSize); err != nil {
		return err
	}

	vaultUtxos, err := GetVaultUtxos(depositTxs, vaultPkScript)
	if err != nil {
		return err
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

		if txIn.Sequence != LiquidationCetSequence {
			return errorsmod.Wrap(ErrInvalidCET, "invalid sequence")
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

	borrowerAuthPubKeyBytes, err := hex.DecodeString(borrowerAuthPubKey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode borrower auth public key")
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

		if !adaptor.Verify(sigBytes, sigHash, borrowerAuthPubKeyBytes, adaptorPoint) {
			return ErrInvalidAdaptorSignature
		}
	}

	return nil
}

// VerifyRepaymentCet verifies the given repayment cet and corresponding signatures
func VerifyRepaymentCet(dlcMeta *DLCMeta, depositTxs []*psbt.Packet, vaultPkScript []byte, borrowerPubKey string, dcmPubKey string, repaymentCet string, signatures []string) error {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(repaymentCet)), true)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidCET, "failed to deserialize cet")
	}

	if len(p.UnsignedTx.TxOut) == 0 {
		return errorsmod.Wrap(ErrInvalidCET, "missing tx out")
	}

	if slices.ContainsFunc(p.UnsignedTx.TxOut, btcbridgetypes.IsDustOut) {
		return errorsmod.Wrap(ErrInvalidCET, "dust tx out")
	}

	script, controlBlock, err := UnwrapLeafScript(dlcMeta.RepaymentScript)
	if err != nil {
		return err
	}

	witnessSize := GetCetWitnessSize(CetType_REPAYMENT, script, controlBlock)

	if err := CheckCetFeeRate(p, witnessSize, 0, 0); err != nil {
		return err
	}

	if err := CheckTransactionWeight(p.UnsignedTx, witnessSize); err != nil {
		return err
	}

	vaultUtxos, err := GetVaultUtxos(depositTxs, vaultPkScript)
	if err != nil {
		return err
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

	borrowerPubKeyBytes, err := hex.DecodeString(borrowerPubKey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode borrower public key")
	}

	for i, signature := range signatures {
		sigHash, err := CalcTapscriptSigHash(p, i, DefaultSigHashType, script)
		if err != nil {
			return errorsmod.Wrapf(err, "failed to calculate sig hash")
		}

		sigBytes, err := hex.DecodeString(signature)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidSignature, "failed to decode signature")
		}

		if !schnorr.Verify(sigBytes, sigHash, borrowerPubKeyBytes) {
			return ErrInvalidSignature
		}
	}

	return nil
}

// CreateTimeoutRefundTransaction creates the timeout refund tx
func CreateTimeoutRefundTransaction(depositTxs []*psbt.Packet, vaultPkScript []byte, borrowerPkScript []byte, internalKeyBytes []byte, leafScript LeafScript, lockTime int64, feeRate int64) (string, error) {
	vaultUtxos, err := GetVaultUtxos(depositTxs, vaultPkScript)
	if err != nil {
		return "", err
	}

	script, controlBlock, err := UnwrapLeafScript(leafScript)
	if err != nil {
		return "", err
	}

	p, err := BuildPsbt(vaultUtxos, borrowerPkScript, feeRate, 64+len(script)+len(controlBlock))
	if err != nil {
		return "", err
	}

	p.UnsignedTx.LockTime = uint32(lockTime)

	for i := range p.Inputs {
		p.Inputs[i].SighashType = DefaultSigHashType
		p.Inputs[i].TaprootInternalKey = internalKeyBytes
		p.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: controlBlock,
				Script:       script,
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
func BuildSignedCet(cet string, borrowerPubKey string, borrowerSignatures []string, dcmPubKey string, dcmSignatures []string, cetType CetType) ([]byte, *chainhash.Hash, error) {
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

	borrowerSigHashType, dcmSigHashType := GetCetSigHashTypes(cetType)

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
				SigHash:     dcmSigHashType,
			},
			{
				XOnlyPubKey: borrowerPubKeyBytes,
				LeafHash:    leafHash[:],
				Signature:   borrowerSig,
				SigHash:     borrowerSigHashType,
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

// GetCetSigHashTypes gets the cet sig hash types for borrower and DCM by the given cet type
// NOTE: The sig hash type for all cets is SigHashDefault currently
func GetCetSigHashTypes(cetType CetType) (txscript.SigHashType, txscript.SigHashType) {
	switch cetType {
	default:
		return DefaultSigHashType, DefaultSigHashType
	}
}

// GetCet gets the liquidation cet and corresponding type according to the given loan status
func GetLiquidationCetAndType(dlcMeta *DLCMeta, loanStatus LoanStatus) (LiquidationCet, CetType) {
	switch loanStatus {
	case LoanStatus_Liquidated:
		return dlcMeta.LiquidationCet, CetType_LIQUIDATION

	default:
		return dlcMeta.DefaultLiquidationCet, CetType_DEFAULT_LIQUIDATION
	}
}

// UpdateLiquidationCet updates the liquidation cet by the given type
func UpdateLiquidationCet(dlcMeta *DLCMeta, cetType CetType, cet LiquidationCet) {
	switch cetType {
	case CetType_LIQUIDATION:
		dlcMeta.LiquidationCet = cet

	default:
		dlcMeta.DefaultLiquidationCet = cet
	}
}

// GetCetInfo gets the cet info from the given event and script
func GetCetInfo(event *dlctypes.DLCEvent, outcomeIndex int, script []byte, controlBlock []byte, sigHashType txscript.SigHashType) (*CetInfo, error) {
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
		Script:         GetLeafScript(script, controlBlock),
		SighashType:    uint32(sigHashType),
	}, nil
}

// GetCetSigHashes gets the sig hashes by the given cet type
func GetCetSigHashes(dlcMeta *DLCMeta, cetType CetType) ([]string, error) {
	var cet string
	var script string

	switch cetType {
	case CetType_LIQUIDATION:
		cet = dlcMeta.LiquidationCet.Tx
		script = dlcMeta.LiquidationScript.Script

	case CetType_DEFAULT_LIQUIDATION:
		cet = dlcMeta.DefaultLiquidationCet.Tx
		script = dlcMeta.LiquidationScript.Script

	case CetType_REPAYMENT:
		cet = dlcMeta.RepaymentCet.Tx
		script = dlcMeta.RepaymentScript.Script
	}

	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(cet)), true)
	if err != nil {
		return nil, err
	}

	scriptBytes, err := hex.DecodeString(script)
	if err != nil {
		return nil, err
	}

	_, sigHashType := GetCetSigHashTypes(cetType)

	sigHashes := []string{}

	for i := range p.Inputs {
		sigHash, err := CalcTapscriptSigHash(p, i, sigHashType, scriptBytes)
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

// GetLeafScript gets a leaf script from the given script and control block
func GetLeafScript(script []byte, controlBlock []byte) LeafScript {
	return LeafScript{
		Script:       hex.EncodeToString(script),
		ControlBlock: hex.EncodeToString(controlBlock),
	}
}

// UnwrapLeafScript unwraps the given leaf script
func UnwrapLeafScript(leafScript LeafScript) ([]byte, []byte, error) {
	script, err := hex.DecodeString(leafScript.Script)
	if err != nil {
		return nil, nil, err
	}

	controlBlock, err := hex.DecodeString(leafScript.ControlBlock)
	if err != nil {
		return nil, nil, err
	}

	return script, controlBlock, nil
}

// CheckCetFeeRate checks the fee rate of the given cet
func CheckCetFeeRate(p *psbt.Packet, witnessSize int, currentFeeRate int64, maxFeeRateMultiplier int64) error {
	virtualSize := GetTxVirtualSize(p.UnsignedTx, witnessSize)

	fee, err := p.GetTxFee()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidCET, "failed to get tx fee: %v", err)
	}

	if int64(fee) < virtualSize {
		return errorsmod.Wrap(ErrInvalidCET, "too low fee rate")
	}

	if maxFeeRateMultiplier > 0 && int64(fee) > virtualSize*currentFeeRate*maxFeeRateMultiplier {
		return errorsmod.Wrap(ErrInvalidCET, "too high fee rate")
	}

	return nil
}

// GetCetWitnessSize gets the cet witness size according to the given params
// NOTE: The final signature is 64 bytes due to that the sig hash type is SigHashDefault currently.
func GetCetWitnessSize(cetType CetType, script []byte, controlBlock []byte) int {
	switch cetType {
	default:
		// dcm signature(64) + borrower signature(64) + len(script) + len(control block)
		return 64 + 64 + len(script) + len(controlBlock)
	}
}

// ParseDepositTxs parses the given deposit txs
// Assume that the given deposit txs are valid psbts
func ParseDepositTxs(depositTxs []string, vaultPkScript []byte) ([]*psbt.Packet, []string, sdkmath.Int, error) {
	parsedDepositTxs := []*psbt.Packet{}
	depositTxHashes := []string{}
	depositAmount := sdkmath.ZeroInt()

	for _, depositTx := range depositTxs {
		p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(depositTx)), true)

		parsedDepositTxs = append(parsedDepositTxs, p)
		depositTxHashes = append(depositTxHashes, p.UnsignedTx.TxHash().String())

		vaultFound := false

		for _, out := range p.UnsignedTx.TxOut {
			if bytes.Equal(out.PkScript, vaultPkScript) {
				depositAmount = depositAmount.Add(sdkmath.NewInt(out.Value))
				vaultFound = true
			}
		}

		if !vaultFound {
			return nil, nil, sdkmath.Int{}, errorsmod.Wrapf(ErrInvalidDepositTx, "no vault found in the deposit tx %s", p.UnsignedTx.TxHash().String())
		}
	}

	return parsedDepositTxs, depositTxHashes, depositAmount, nil
}

// GetVaultUtxos gets the vault utxos from the given deposit txs
func GetVaultUtxos(depositTxs []*psbt.Packet, vaultPkScript []byte) ([]*btcbridgetypes.UTXO, error) {
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
		return nil, errorsmod.Wrapf(ErrInvalidDepositTx, "no vault found in the deposit tx %s", depositTx.UnsignedTx.TxHash().String())
	}

	return utxos, nil
}
