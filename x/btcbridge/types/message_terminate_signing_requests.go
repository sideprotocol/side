package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgTerminateSigningRequests = "terminate_signing_requests"

func (msg *MsgTerminateSigningRequests) Route() string {
	return RouterKey
}

func (msg *MsgTerminateSigningRequests) Type() string {
	return TypeMsgTerminateSigningRequests
}

func (msg *MsgTerminateSigningRequests) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

func (msg *MsgTerminateSigningRequests) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgTerminateSigningRequests) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid sender address (%s)", err)
	}

	return nil
}
