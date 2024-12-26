package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgFund{}

func NewMsgSubmitFundingTx(relayer string, fundingTx string, height uint64, proof string) *MsgFund {
	return &MsgFund{
		Relayer:   relayer,
		FundingTx: fundingTx,
		Height:    height,
		Poof:      proof,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgFund) ValidateBasic() error {
	if len(m.FundingTx) == 0 {
		return ErrEmptyFundTx
	}

	if len(m.Poof) == 0 {
		return ErrEmptyPoof
	}

	return nil
}
