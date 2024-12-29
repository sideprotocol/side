package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgApprove{}

func NewMsgDeposit(relayer string, DepositTxId string, height uint64, proof string) *MsgApprove {
	return &MsgApprove{
		Relayer:     relayer,
		DepositTxId: DepositTxId,
		Height:      height,
		Poof:        proof,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgApprove) ValidateBasic() error {
	if len(m.DepositTxId) == 0 {
		return ErrEmptyDepositTx
	}

	if len(m.Poof) == 0 {
		return ErrEmptyPoof
	}

	return nil
}
