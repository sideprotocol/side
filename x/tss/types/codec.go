package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCompleteDKG{}, "tss/MsgCompleteDKG", nil)
	cdc.RegisterConcrete(&MsgSubmitSignatures{}, "tss/MsgSubmitSignatures", nil)
	cdc.RegisterConcrete(&MsgRefresh{}, "tss/MsgRefresh", nil)
	cdc.RegisterConcrete(&MsgCompleteRefreshing{}, "tss/MsgCompleteRefreshing", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "tss/MsgUpdateParams", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCompleteDKG{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitSignatures{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRefresh{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCompleteRefreshing{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateParams{})

	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
