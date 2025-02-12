package types

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgApprove{}

func NewMsgApprove(relayer string, depositTxId string, blockHash string, proof []string) *MsgApprove {
	return &MsgApprove{
		Relayer:     relayer,
		DepositTxId: depositTxId,
		BlockHash:   blockHash,
		Proof:       proof,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgApprove) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Relayer); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if _, err := chainhash.NewHashFromStr(m.DepositTxId); err != nil {
		return ErrInvalidDepositTx
	}

	if _, err := chainhash.NewHashFromStr(m.BlockHash); err != nil {
		return ErrInvalidBlockHash
	}

	if len(m.Proof) == 0 {
		return ErrInvalidProof
	}

	return nil
}
