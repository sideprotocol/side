package types

import "github.com/btcsuite/btcd/txscript"

const (
	// tag for the IBC channel through which the ibc transfer will be performed
	TagChannel = txscript.OP_0

	// tag for the recipient which represents the recipient address on the destination chain
	TagRecipient = txscript.OP_1

	// tag for auto pegout enabling
	TagAutoPegout = "auto-pegout"

	// default max gas for IBC callback
	DefaultMaxIBCCallbackGas = uint64(1_000_000)
)

// IBCTransferMaybeEnabled returns true if the deposit script maybe indicates enabling the IBC transfer, false otherwise
func IBCTransferMaybeEnabled(depositScript []byte) bool {
	return txscript.IsNullData(depositScript)
}

// ParseIBCTransfer parses the channel id and recipient address from the given deposit script
func ParseIBCTransfer(depositScript []byte) (channelId string, recipient string, err error) {
	tokenizer := txscript.MakeScriptTokenizer(0, depositScript)
	if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != txscript.OP_RETURN {
		return "", "", ErrInvalidDepositScript
	}

	if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != TagChannel {
		return "", "", ErrInvalidDepositScript
	}

	channelId = string(tokenizer.Data())

	if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != TagRecipient {
		return "", "", ErrInvalidDepositScript
	}

	recipient = string(tokenizer.Data())

	if tokenizer.Next() {
		return "", "", ErrInvalidDepositScript
	}

	return
}
