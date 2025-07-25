package types

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	errorsmod "cosmossdk.io/errors"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
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
	// unprefix channel id
	unprefixedChannelId, err := GetUnprefixedChannelId(channelId)
	if err != nil {
		return nil, err
	}

	// build OP_RETURN script
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddOp(txscript.OP_RETURN)

	// add magic number
	scriptBuilder.AddOp(IBCTransferMagicNumber)

	// add payload
	scriptBuilder.AddData(unprefixedChannelId).AddData([]byte(recipient))

	return scriptBuilder.Script()
}

// GetIBCTransferScript gets the IBC transfer script from the given deposit tx
func GetIBCTransferScript(depositTx *wire.MsgTx) []byte {
	for _, out := range depositTx.TxOut {
		if IsOpReturnOutput(out) && len(out.PkScript) > 1 && out.PkScript[1] == IBCTransferMagicNumber {
			return out.PkScript
		}
	}

	return nil
}

// ParseIBCTransferScript parses the channel id and recipient address from the given script
func ParseIBCTransferScript(script []byte) (channelId string, recipient string, err error) {
	tokenizer := txscript.MakeScriptTokenizer(0, script)
	if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != txscript.OP_RETURN {
		return "", "", errorsmod.Wrap(ErrInvalidIBCTransferScript, "non OP_RETURN script")
	}

	if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != IBCTransferMagicNumber {
		return "", "", errorsmod.Wrap(ErrInvalidIBCTransferScript, "failed to parse magic number")
	}

	if !tokenizer.Next() || tokenizer.Err() != nil {
		return "", "", errorsmod.Wrap(ErrInvalidIBCTransferScript, "failed to parse channel id")
	}

	if len(tokenizer.Data()) != 4 {
		return "", "", errorsmod.Wrap(ErrInvalidIBCTransferScript, "invalid channel id")
	}

	channelId = NormalizeChannelId(tokenizer.Data())

	if !tokenizer.Next() || tokenizer.Err() != nil {
		return "", "", errorsmod.Wrap(ErrInvalidIBCTransferScript, "failed to parse recipient address")
	}

	recipient = string(tokenizer.Data())

	if tokenizer.Next() {
		return "", "", ErrInvalidIBCTransferScript
	}

	return
}

// GetUnprefixedChannelId gets the channel id without prefix
func GetUnprefixedChannelId(channelId string) ([]byte, error) {
	unprefixedChannelId, found := strings.CutPrefix(channelId, channeltypes.ChannelPrefix)
	if !found {
		return nil, channeltypes.ErrInvalidChannelIdentifier
	}

	id, err := strconv.ParseUint(unprefixedChannelId, 10, 32)
	if err != nil {
		return nil, err
	}

	bz := make([]byte, 4)
	binary.BigEndian.PutUint32(bz, uint32(id))

	return bz, nil
}

// NormalizeChannelId normalizes the given channel id
func NormalizeChannelId(unprefixedChannelId []byte) string {
	if len(unprefixedChannelId) != 4 {
		return ""
	}

	return fmt.Sprintf("%s%d", channeltypes.ChannelPrefix, binary.BigEndian.Uint32(unprefixedChannelId))
}
