package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgApprove{}

func NewMsgDeposit(relayer string, depositTxId string, blockHash string, proof []string) *MsgApprove {
	return &MsgApprove{
		Relayer:     relayer,
		DepositTxId: depositTxId,
		BlockHash:   blockHash,
		Proof:       proof,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgApprove) ValidateBasic() error {
	if len(m.DepositTxId) == 0 {
		return ErrEmptyDepositTx
	}

	if len(m.Proof) == 0 {
		return ErrInvalidProof
	}

	return nil
}
