package types

import (
	"bytes"
	"encoding/base64"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitDepositTransaction{}

func NewMsgSubmitDepositTransaction(relayer string, vault string, depositTx string, blockHash string, proof []string) *MsgSubmitDepositTransaction {
	return &MsgSubmitDepositTransaction{
		Relayer:   relayer,
		Vault:     vault,
		DepositTx: depositTx,
		BlockHash: blockHash,
		Proof:     proof,
	}
}

// ValidateBasic performs basic MsgSubmitDepositTransaction message validation.
func (m *MsgSubmitDepositTransaction) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Relayer); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	vaultPkScript, err := GetPkScriptFromAddress(m.Vault)
	if err != nil {
		return ErrInvalidVault
	}

	txBytes, err := base64.StdEncoding.DecodeString(m.DepositTx)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidDepositTx, "failed to decode deposit tx")
	}

	var tx wire.MsgTx
	if err := tx.Deserialize(bytes.NewReader(txBytes)); err != nil {
		return errorsmod.Wrap(ErrInvalidDepositTx, "failed to deserialize deposit tx")
	}

	vaultFound := false
	for _, out := range tx.TxOut {
		if bytes.Equal(out.PkScript, vaultPkScript) {
			vaultFound = true
			break
		}
	}

	if !vaultFound {
		return errorsmod.Wrap(ErrInvalidDepositTx, "vault does not exist in tx outs")
	}

	if _, err := chainhash.NewHashFromStr(m.BlockHash); err != nil {
		return ErrInvalidBlockHash
	}

	if len(m.Proof) == 0 {
		return ErrInvalidProof
	}

	return nil
}
