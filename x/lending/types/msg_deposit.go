package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgDeposit{}

func NewMsgDeposit(relayer string, DepositTxId string, height uint64, proof string) *MsgDeposit {
	return &MsgDeposit{
		Relayer:     relayer,
		DepositTxId: DepositTxId,
		Height:      height,
		Poof:        proof,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgDeposit) ValidateBasic() error {
	if len(m.DepositTxId) == 0 {
		return ErrEmptyDepositTx
	}

	if len(m.Poof) == 0 {
		return ErrEmptyPoof
	}

	return nil
}
