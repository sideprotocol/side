package types

import (
	"lukechampine.com/uint128"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// default tx version
	TxVersion = 2

	// default minimum relay fee
	MinRelayFee = 1000

	// default maximum allowed transaction weight
	MaxTransactionWeight = 400000

	// default sig hash type
	DefaultSigHashType = txscript.SigHashDefault
)

// BuildPsbt builds a bitcoin psbt from the given params.
// Assume that the utxo script type is witness.
func BuildPsbt(utxoIterator UTXOIterator, recipient string, amount int64, feeRate int64, change string, maxUTXONum int) (*psbt.Packet, []*UTXO, *UTXO, error) {
	chaincfg := sdk.GetConfig().GetBtcChainCfg()

	recipientAddr, err := btcutil.DecodeAddress(recipient, chaincfg)
	if err != nil {
		return nil, nil, nil, err
	}

	recipientPkScript, err := txscript.PayToAddrScript(recipientAddr)
	if err != nil {
		return nil, nil, nil, err
	}

	changeAddr, err := btcutil.DecodeAddress(change, chaincfg)
	if err != nil {
		return nil, nil, nil, err
	}

	txOuts := make([]*wire.TxOut, 0)
	txOuts = append(txOuts, wire.NewTxOut(amount, recipientPkScript))

	unsignedTx, selectedUTXOs, changeUTXO, err := BuildUnsignedTransaction([]*UTXO{}, txOuts, utxoIterator, feeRate, changeAddr, maxUTXONum)
	if err != nil {
		return nil, nil, nil, err
	}

	p, err := psbt.NewFromUnsignedTx(unsignedTx)
	if err != nil {
		return nil, nil, nil, err
	}

	for i, utxo := range selectedUTXOs {
		p.Inputs[i].SighashType = DefaultSigHashType
		p.Inputs[i].WitnessUtxo = wire.NewTxOut(int64(utxo.Amount), utxo.PubKeyScript)
	}

	return p, selectedUTXOs, changeUTXO, nil
}

// BuildTransferAllBtcPsbt builds a bitcoin psbt to transfer all given btc.
// Assume that the utxo script type is witness.
func BuildTransferAllBtcPsbt(utxos []*UTXO, recipient string, feeRate int64) (*psbt.Packet, *UTXO, error) {
	chaincfg := sdk.GetConfig().GetBtcChainCfg()

	recipientAddr, err := btcutil.DecodeAddress(recipient, chaincfg)
	if err != nil {
		return nil, nil, err
	}

	recipientPkScript, err := txscript.PayToAddrScript(recipientAddr)
	if err != nil {
		return nil, nil, err
	}

	txOuts := make([]*wire.TxOut, 0)
	txOuts = append(txOuts, wire.NewTxOut(0, recipientPkScript))

	unsignedTx, err := BuildUnsignedTransactionWithoutExtraChange([]*UTXO{}, txOuts, utxos, feeRate)
	if err != nil {
		return nil, nil, err
	}

	p, err := psbt.NewFromUnsignedTx(unsignedTx)
	if err != nil {
		return nil, nil, err
	}

	for i, utxo := range utxos {
		p.Inputs[i].SighashType = DefaultSigHashType
		p.Inputs[i].WitnessUtxo = wire.NewTxOut(int64(utxo.Amount), utxo.PubKeyScript)
	}

	recipientUTXO := GetChangeUTXO(unsignedTx, recipient)

	return p, recipientUTXO, nil
}

// BuildBtcBatchWithdrawPsbt builds the psbt to perform btc batch withdrawal
func BuildBtcBatchWithdrawPsbt(utxoIterator UTXOIterator, withdrawRequests []*WithdrawRequest, feeRate int64, change string, maxUTXONum int) (*psbt.Packet, []*UTXO, *UTXO, error) {
	chainCfg := sdk.GetConfig().GetBtcChainCfg()

	txOuts := make([]*wire.TxOut, len(withdrawRequests))

	for i, req := range withdrawRequests {
		amount, _ := sdk.ParseCoinNormalized(req.Amount)

		address, err := btcutil.DecodeAddress(req.Address, chainCfg)
		if err != nil {
			return nil, nil, nil, err
		}

		pkScript, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, nil, nil, err
		}

		txOuts[i] = wire.NewTxOut(int64(amount.Amount.Uint64()), pkScript)
	}

	changeAddress, err := btcutil.DecodeAddress(change, chainCfg)
	if err != nil {
		return nil, nil, nil, err
	}

	unsignedTx, selectedUTXOs, changeUTXO, err := BuildUnsignedTransaction([]*UTXO{}, txOuts, utxoIterator, feeRate, changeAddress, maxUTXONum)
	if err != nil {
		return nil, nil, nil, err
	}

	p, err := psbt.NewFromUnsignedTx(unsignedTx)
	if err != nil {
		return nil, nil, nil, err
	}

	for i, utxo := range selectedUTXOs {
		p.Inputs[i].SighashType = DefaultSigHashType
		p.Inputs[i].WitnessUtxo = wire.NewTxOut(int64(utxo.Amount), utxo.PubKeyScript)
	}

	return p, selectedUTXOs, changeUTXO, nil
}

// BuildRunesPsbt builds a bitcoin psbt for runes edict from the given params.
// Assume that the utxo script type is witness.
func BuildRunesPsbt(utxos []*UTXO, paymentUTXOIterator UTXOIterator, recipient string, runeId string, amount uint128.Uint128, feeRate int64, runeBalancesDelta []*RuneBalance, runesChange string, change string, maxUTXONum int) (*psbt.Packet, []*UTXO, *UTXO, *UTXO, error) {
	chaincfg := sdk.GetConfig().GetBtcChainCfg()

	recipientAddr, err := btcutil.DecodeAddress(recipient, chaincfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	recipientPkScript, err := txscript.PayToAddrScript(recipientAddr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	changeAddr, err := btcutil.DecodeAddress(change, chaincfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	runesChangeAddr, err := btcutil.DecodeAddress(runesChange, chaincfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	runesChangePkScript, err := txscript.PayToAddrScript(runesChangeAddr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	txOuts := make([]*wire.TxOut, 0)

	// fill the runes protocol script with empty output script first
	txOuts = append(txOuts, wire.NewTxOut(0, []byte{}))

	var runesChangeUTXO *UTXO
	edictOutputIndex := uint32(1)

	if len(runeBalancesDelta) > 0 {
		runesChangeUTXO = GetRunesChangeUTXO(runeBalancesDelta, runesChange, runesChangePkScript, 1)

		// allocate the remaining runes to the first non-OP_RETURN output by default
		txOuts = append(txOuts, wire.NewTxOut(RunesOutValue, runesChangePkScript))

		// advance the edict output index
		edictOutputIndex++
	}

	// edict output
	txOuts = append(txOuts, wire.NewTxOut(RunesOutValue, recipientPkScript))

	runesScript, err := BuildEdictScript(runeId, amount, edictOutputIndex)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// populate the runes protocol script
	txOuts[0].PkScript = runesScript

	unsignedTx, selectedUTXOs, changeUTXO, err := BuildUnsignedTransaction(utxos, txOuts, paymentUTXOIterator, feeRate, changeAddr, maxUTXONum)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if runesChangeUTXO != nil {
		runesChangeUTXO.Txid = unsignedTx.TxHash().String()
	}

	p, err := psbt.NewFromUnsignedTx(unsignedTx)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for i, utxo := range utxos {
		p.Inputs[i].SighashType = DefaultSigHashType
		p.Inputs[i].WitnessUtxo = wire.NewTxOut(int64(utxo.Amount), utxo.PubKeyScript)
	}

	for i, utxo := range selectedUTXOs {
		p.Inputs[i+len(utxos)].SighashType = DefaultSigHashType
		p.Inputs[i+len(utxos)].WitnessUtxo = wire.NewTxOut(int64(utxo.Amount), utxo.PubKeyScript)
	}

	return p, selectedUTXOs, changeUTXO, runesChangeUTXO, nil
}

// BuildTransferAllRunesPsbt builds a bitcoin psbt to transfer all specified runes.
// Assume that the utxo script type is witness.
func BuildTransferAllRunesPsbt(utxos []*UTXO, paymentUTXOIterator UTXOIterator, recipient string, runeBalancesDelta []*RuneBalance, feeRate int64, btcChange string, maxUTXONum int) (*psbt.Packet, []*UTXO, *UTXO, *UTXO, error) {
	chaincfg := sdk.GetConfig().GetBtcChainCfg()

	recipientAddr, err := btcutil.DecodeAddress(recipient, chaincfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	recipientPkScript, err := txscript.PayToAddrScript(recipientAddr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	changeAddr, err := btcutil.DecodeAddress(btcChange, chaincfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	txOuts := make([]*wire.TxOut, 0)

	// fill the runes protocol script without payload
	txOuts = append(txOuts, wire.NewTxOut(0, []byte{txscript.OP_RETURN, txscript.OP_13}))

	// allocate the remaining runes to the first non-OP_RETURN output by default
	txOuts = append(txOuts, wire.NewTxOut(RunesOutValue, recipientPkScript))

	unsignedTx, selectedUTXOs, changeUTXO, err := BuildUnsignedTransaction(utxos, txOuts, paymentUTXOIterator, feeRate, changeAddr, maxUTXONum)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	p, err := psbt.NewFromUnsignedTx(unsignedTx)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	runesRecipientUTXO := GetRunesChangeUTXO(runeBalancesDelta, recipient, recipientPkScript, 1)
	runesRecipientUTXO.Txid = unsignedTx.TxHash().String()

	for i, utxo := range utxos {
		p.Inputs[i].SighashType = DefaultSigHashType
		p.Inputs[i].WitnessUtxo = wire.NewTxOut(int64(utxo.Amount), utxo.PubKeyScript)
	}

	for i, utxo := range selectedUTXOs {
		p.Inputs[i+len(utxos)].SighashType = DefaultSigHashType
		p.Inputs[i+len(utxos)].WitnessUtxo = wire.NewTxOut(int64(utxo.Amount), utxo.PubKeyScript)
	}

	return p, selectedUTXOs, changeUTXO, runesRecipientUTXO, nil
}

// BuildUnsignedTransaction builds an unsigned tx from the given params.
func BuildUnsignedTransaction(utxos []*UTXO, txOuts []*wire.TxOut, paymentUTXOIterator UTXOIterator, feeRate int64, change btcutil.Address, maxUTXONum int) (*wire.MsgTx, []*UTXO, *UTXO, error) {
	tx := wire.NewMsgTx(TxVersion)

	inAmount := int64(0)
	outAmount := int64(0)

	for _, utxo := range utxos {
		AddUTXOToTx(tx, utxo)
		inAmount += int64(utxo.Amount)
	}

	for _, txOut := range txOuts {
		if IsDustOut(txOut) {
			return nil, nil, nil, ErrDustOutput
		}

		tx.AddTxOut(txOut)
		outAmount += txOut.Value
	}

	changePkScript, err := txscript.PayToAddrScript(change)
	if err != nil {
		return nil, nil, nil, err
	}

	changeOut := wire.NewTxOut(0, changePkScript)

	selectedUTXOs, err := AddPaymentUTXOsToTx(tx, utxos, inAmount-outAmount, paymentUTXOIterator, changeOut, feeRate, maxUTXONum)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := CheckTransactionWeight(tx, append(utxos, selectedUTXOs...)); err != nil {
		return nil, nil, nil, err
	}

	var changeUTXO *UTXO
	if len(tx.TxOut) > len(txOuts) {
		changeUTXO = GetChangeUTXO(tx, change.EncodeAddress())
	}

	return tx, selectedUTXOs, changeUTXO, nil
}

// BuildUnsignedTransactionWithoutExtraChange builds an unsigned tx from the given params.
// All payment utxos will be spent and the last out is the recipient (and change) out.
func BuildUnsignedTransactionWithoutExtraChange(utxos []*UTXO, txOuts []*wire.TxOut, paymentUTXOs []*UTXO, feeRate int64) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(TxVersion)

	inAmount := int64(0)
	outAmount := int64(0)

	for _, utxo := range utxos {
		AddUTXOToTx(tx, utxo)
		inAmount += int64(utxo.Amount)
	}

	for _, utxo := range paymentUTXOs {
		AddUTXOToTx(tx, utxo)
		inAmount += int64(utxo.Amount)
	}

	for i, txOut := range txOuts {
		if i != len(txOuts)-1 && IsDustOut(txOut) {
			return nil, ErrDustOutput
		}

		tx.AddTxOut(txOut)
		outAmount += txOut.Value
	}

	fee := GetTxVirtualSize(tx, append(utxos, paymentUTXOs...)) * feeRate

	change := inAmount - outAmount - fee
	if change <= 0 {
		return nil, ErrInsufficientUTXOs
	}

	txOuts[len(txOuts)-1].Value += change
	if IsDustOut(txOuts[len(txOuts)-1]) {
		return nil, ErrDustOutput
	}

	if err := CheckTransactionWeight(tx, append(utxos, paymentUTXOs...)); err != nil {
		return nil, err
	}

	return tx, nil
}

// AddPaymentUTXOsToTx adds the given payment utxos to the tx.
func AddPaymentUTXOsToTx(tx *wire.MsgTx, utxos []*UTXO, inOutDiff int64, paymentUTXOIterator UTXOIterator, changeOut *wire.TxOut, feeRate int64, maxUTXONum int) ([]*UTXO, error) {
	selectedUTXOs := make([]*UTXO, 0)
	paymentValue := int64(0)

	defer paymentUTXOIterator.Close()

	for ; paymentUTXOIterator.Valid(); paymentUTXOIterator.Next() {
		utxo := paymentUTXOIterator.GetUTXO()
		if utxo.IsLocked {
			continue
		}

		utxos = append(utxos, utxo)
		if maxUTXONum != 0 && len(utxos) > maxUTXONum {
			return nil, ErrMaxUTXONumExceeded
		}

		selectedUTXOs = append(selectedUTXOs, utxo)

		AddUTXOToTx(tx, utxo)
		tx.AddTxOut(changeOut)

		paymentValue += int64(utxo.Amount)
		fee := GetTxVirtualSize(tx, utxos) * feeRate

		changeValue := paymentValue + inOutDiff - fee
		if changeValue > 0 {
			tx.TxOut[len(tx.TxOut)-1].Value = changeValue
			if IsDustOut(tx.TxOut[len(tx.TxOut)-1]) {
				tx.TxOut = tx.TxOut[0 : len(tx.TxOut)-1]
			}

			return selectedUTXOs, nil
		}

		tx.TxOut = tx.TxOut[0 : len(tx.TxOut)-1]

		if changeValue == 0 {
			return selectedUTXOs, nil
		}

		if changeValue < 0 {
			feeWithoutChange := GetTxVirtualSize(tx, utxos) * feeRate
			if paymentValue+inOutDiff-feeWithoutChange >= 0 {
				return selectedUTXOs, nil
			}
		}
	}

	return nil, ErrInsufficientUTXOs
}

// AddUTXOToTx adds the given utxo to the specified tx
// Make sure the utxo is valid
func AddUTXOToTx(tx *wire.MsgTx, utxo *UTXO) {
	txIn := new(wire.TxIn)

	hash, err := chainhash.NewHashFromStr(utxo.Txid)
	if err != nil {
		panic(err)
	}

	txIn.PreviousOutPoint = *wire.NewOutPoint(hash, uint32(utxo.Vout))
	txIn.Sequence = MagicSequence

	tx.AddTxIn(txIn)
}

// GetChangeUTXO returns the change output from the given tx
// Make sure that the tx is valid and the change output is the last output
func GetChangeUTXO(tx *wire.MsgTx, change string) *UTXO {
	changeOut := tx.TxOut[len(tx.TxOut)-1]

	return &UTXO{
		Txid:         tx.TxHash().String(),
		Vout:         uint64(len(tx.TxOut) - 1),
		Address:      change,
		Amount:       uint64(changeOut.Value),
		PubKeyScript: changeOut.PkScript,
	}
}

// GetRunesChangeUTXO gets the runes change utxo.
func GetRunesChangeUTXO(runeBalancesDelta []*RuneBalance, change string, changePkScript []byte, outIndex uint32) *UTXO {
	return &UTXO{
		Vout:         uint64(outIndex),
		Address:      change,
		Amount:       RunesOutValue,
		PubKeyScript: changePkScript,
		Runes:        runeBalancesDelta,
	}
}

// GetTxVirtualSize gets the virtual size of the given tx.
func GetTxVirtualSize(tx *wire.MsgTx, utxos []*UTXO) int64 {
	newTx := PopulateTxWithDummyWitness(tx, utxos)

	return mempool.GetTxVirtualSize(btcutil.NewTx(newTx))
}

// CheckTransactionWeight checks if the weight of the given tx exceeds the allowed maximum weight
func CheckTransactionWeight(tx *wire.MsgTx, utxos []*UTXO) error {
	newTx := PopulateTxWithDummyWitness(tx, utxos)

	weight := blockchain.GetTransactionWeight(btcutil.NewTx(newTx))
	if weight > MaxTransactionWeight {
		return ErrMaxTransactionWeightExceeded
	}

	return nil
}

// PopulateTxWithDummyWitness populates the given tx with the dummy witness
// Assume that the utxo script type is the witness type.
// If utxos are not provided, the witness type is defaulted to taproot
func PopulateTxWithDummyWitness(tx *wire.MsgTx, utxos []*UTXO) *wire.MsgTx {
	if len(utxos) == 0 {
		return PopulateTxWithDummyTaprootWitness(tx)
	}

	newTx := tx.Copy()

	for i, txIn := range newTx.TxIn {
		var dummyWitness []byte

		switch txscript.GetScriptClass(utxos[i].PubKeyScript) {
		case txscript.WitnessV1TaprootTy:
			// maximum size when the sig hash is not SigHashDefault
			dummyWitness = make([]byte, 65)

		case txscript.WitnessV0PubKeyHashTy:
			dummyWitness = make([]byte, 73+33)

		default:
		}

		txIn.Witness = wire.TxWitness{dummyWitness}
	}

	return newTx
}

// PopulateTxWithDummyTaprootWitness populates the given tx with the dummy taproot witness
func PopulateTxWithDummyTaprootWitness(tx *wire.MsgTx) *wire.MsgTx {
	newTx := tx.Copy()

	for _, txIn := range newTx.TxIn {
		// maximum size when the sig hash is not SigHashDefault
		dummyWitness := make([]byte, 65)

		txIn.Witness = wire.TxWitness{dummyWitness}
	}

	return newTx
}

// IsDustOut returns true if the given output is dust, false otherwise
func IsDustOut(out *wire.TxOut) bool {
	return !IsOpReturnOutput(out) && mempool.IsDust(out, MinRelayFee)
}

// IsOpReturnOutput returns true if the script of the given out starts with OP_RETURN
func IsOpReturnOutput(out *wire.TxOut) bool {
	return len(out.PkScript) > 0 && out.PkScript[0] == txscript.OP_RETURN
}

// IsValidBtcAddress returns true if the given address is a standard bitcoin address, false otherwise
func IsValidBtcAddress(address string) bool {
	_, err := btcutil.DecodeAddress(address, sdk.GetConfig().GetBtcChainCfg())
	return err == nil
}

// MustPkScriptFromAddress returns the public key script of the given address
// Panic if any error occurred
func MustPkScriptFromAddress(address string) []byte {
	addr, err := btcutil.DecodeAddress(address, sdk.GetConfig().GetBtcChainCfg())
	if err != nil {
		panic(err)
	}

	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		panic(err)
	}

	return pkScript
}
