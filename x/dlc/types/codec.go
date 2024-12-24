package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitNonce{}, "dlc/MsgSubmitNonce", nil)
	cdc.RegisterConcrete(&MsgSubmitAttestation{}, "dlc/MsgSubmitAttestation", nil)
	cdc.RegisterConcrete(&MsgSubmitOraclePubKey{}, "dlc/MsgSubmitOraclePubKey", nil)
	cdc.RegisterConcrete(&MsgSubmitAgencyPubKey{}, "dlc/MsgSubmitAgencyPubKey", nil)
	cdc.RegisterConcrete(&MsgCreateOracle{}, "dlc/MsgCreateOracle", nil)
	cdc.RegisterConcrete(&MsgCreateAgency{}, "dlc/MsgCreateAgency", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "dlc/MsgUpdateParams", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitNonce{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitAttestation{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitOraclePubKey{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitAgencyPubKey{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCreateOracle{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCreateAgency{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateParams{})

	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
