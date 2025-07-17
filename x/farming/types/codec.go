package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgStake{}, "farming/MsgStake", nil)
	cdc.RegisterConcrete(&MsgUnstake{}, "farming/MsgUnstake", nil)
	cdc.RegisterConcrete(&MsgClaim{}, "farming/MsgClaim", nil)
	cdc.RegisterConcrete(&MsgClaimAll{}, "farming/MsgClaimAll", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "farming/MsgUpdateParams", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgStake{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUnstake{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgClaim{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgClaimAll{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateParams{})

	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
