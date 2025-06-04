package types

import (
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	// magic number
	IBCTransferMagicNumber = txscript.OP_10

	// default memo for IBC transfer
	DefaultMemo = "BTC bridge | Side Chain"

	// callback address to enable auto pegout
	CallbackAddress = "btcbridge"

	// default max gas for IBC callback
	DefaultMaxIBCCallbackGas = uint64(1_000_000)
)

// BuildIBCTransferScript builds the script for IBC transfer with the given channel and recipient address
func BuildIBCTransferScript(channelId string, recipient string) ([]byte, error) {
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddOp(txscript.OP_RETURN)

	// add magic number
	scriptBuilder.AddOp(IBCTransferMagicNumber)

	// add payload
	scriptBuilder.AddData([]byte(channelId)).AddData([]byte(recipient))

	return scriptBuilder.Script()
}

// GetIBCTransferScript gets the IBC transfer script from the given deposit tx
func GetIBCTransferScript(depositTx *wire.MsgTx) []byte {
	for _, out := range depositTx.TxOut {
		if IsOpReturnOutput(out) && out.PkScript[1] == IBCTransferMagicNumber {
			return out.PkScript
		}
	}

	return nil
}

// ParseIBCTransferScript parses the channel id and recipient address from the given script
func ParseIBCTransferScript(script []byte) (channelId string, recipient string, err error) {
	tokenizer := txscript.MakeScriptTokenizer(0, script)
	if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != txscript.OP_RETURN {
		return "", "", ErrInvalidIBCTransferScript
	}

	if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != IBCTransferMagicNumber {
		return "", "", ErrInvalidIBCTransferScript
	}

	if !tokenizer.Next() || tokenizer.Err() != nil {
		return "", "", ErrInvalidIBCTransferScript
	}

	channelId = string(tokenizer.Data())

	if !tokenizer.Next() || tokenizer.Err() != nil {
		return "", "", ErrInvalidIBCTransferScript
	}

	recipient = string(tokenizer.Data())

	if tokenizer.Next() {
		return "", "", ErrInvalidIBCTransferScript
	}

	return
}
