package types

import (
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	// magic number
	IBCTransferMagicNumber = txscript.OP_10

	// default port id
	DefaultPortId = "transfer"

	// default memo for IBC transfer
	DefaultMemo = "BTC bridge | Side Chain"

	// flag to enable auto pegout
	FlagAutoPegOut = "auto-pegout"

	// default max gas for IBC callback
	DefaultMaxIBCCallbackGas = uint64(1_000_000)
)

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
