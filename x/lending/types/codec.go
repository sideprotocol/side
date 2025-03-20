package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreatePool{}, "lending/MsgCreatePool", nil)
	cdc.RegisterConcrete(&MsgAddLiquidity{}, "lending/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(&MsgRemoveLiquidity{}, "lending/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(&MsgApply{}, "lending/MsgApply", nil)
	cdc.RegisterConcrete(&MsgSubmitCets{}, "lending/MsgSubmitCets", nil)
	cdc.RegisterConcrete(&MsgApprove{}, "lending/MsgApprove", nil)
	cdc.RegisterConcrete(&MsgSubmitRepaymentAdaptorSignatures{}, "lending/MsgSubmitRepaymentAdaptorSignatures", nil)
	cdc.RegisterConcrete(&MsgCancel{}, "lending/MsgCancel", nil)
	cdc.RegisterConcrete(&MsgSubmitCancellationSignatures{}, "lending/MsgSubmitCancellationSignatures", nil)
	cdc.RegisterConcrete(&MsgRepay{}, "lending/MsgRepay", nil)
	cdc.RegisterConcrete(&MsgSubmitLiquidationSignatures{}, "lending/MsgSubmitLiquidationSignatures", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "lending/MsgUpdateParams", nil)

	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCreatePool{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgAddLiquidity{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRemoveLiquidity{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgApply{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitCets{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgApprove{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitRepaymentAdaptorSignatures{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCancel{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitCancellationSignatures{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRepay{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitLiquidationSignatures{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateParams{})

	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
