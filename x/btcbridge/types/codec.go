package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitBlockHeaders{}, "btcbridge/MsgSubmitBlockHeaders", nil)
	cdc.RegisterConcrete(&MsgSubmitDepositTransaction{}, "btcbridge/MsgSubmitDepositTransaction", nil)
	cdc.RegisterConcrete(&MsgSubmitWithdrawTransaction{}, "btcbridge/MsgSubmitWithdrawTransaction", nil)
	cdc.RegisterConcrete(&MsgSubmitFeeRate{}, "btcbridge/MsgSubmitFeeRate", nil)
	cdc.RegisterConcrete(&MsgUpdateTrustedNonBtcRelayers{}, "btcbridge/MsgUpdateTrustedNonBtcRelayers", nil)
	cdc.RegisterConcrete(&MsgUpdateTrustedOracles{}, "btcbridge/MsgUpdateTrustedOracles", nil)
	cdc.RegisterConcrete(&MsgWithdrawToBitcoin{}, "btcbridge/MsgWithdrawToBitcoin", nil)
	cdc.RegisterConcrete(&MsgSubmitSignatures{}, "btcbridge/MsgSubmitSignatures", nil)
	cdc.RegisterConcrete(&MsgCompleteDKG{}, "btcbridge/MsgCompleteDKG", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "btcbridge/MsgUpdateParams", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitBlockHeaders{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitDepositTransaction{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitWithdrawTransaction{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitFeeRate{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateTrustedNonBtcRelayers{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateTrustedOracles{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgWithdrawToBitcoin{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitSignatures{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCompleteDKG{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateParams{})
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
