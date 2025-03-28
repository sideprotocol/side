package types

import (
	"bytes"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgApprove{}

func NewMsgApprove(relayer string, vault string, depositTx string, blockHash string, proof []string) *MsgApprove {
	return &MsgApprove{
		Relayer:   relayer,
		Vault:     vault,
		DepositTx: depositTx,
		BlockHash: blockHash,
		Proof:     proof,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgApprove) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Relayer); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	vaultPkScript, err := GetPkScriptFromAddress(m.Vault)
	if err != nil {
		return ErrInvalidVault
	}

	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(m.DepositTx)), true)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidDepositTx, "failed to decode deposit tx")
	}

	vaultFound := false
	for _, out := range p.UnsignedTx.TxOut {
		if bytes.Equal(out.PkScript, vaultPkScript) {
			vaultFound = true
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
