// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: side/btcbridge/tx.proto

package btcbridge

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Msg_SubmitBlockHeaders_FullMethodName          = "/side.btcbridge.Msg/SubmitBlockHeaders"
	Msg_SubmitDepositTransaction_FullMethodName    = "/side.btcbridge.Msg/SubmitDepositTransaction"
	Msg_SubmitWithdrawTransaction_FullMethodName   = "/side.btcbridge.Msg/SubmitWithdrawTransaction"
	Msg_SubmitFeeRate_FullMethodName               = "/side.btcbridge.Msg/SubmitFeeRate"
	Msg_UpdateTrustedNonBtcRelayers_FullMethodName = "/side.btcbridge.Msg/UpdateTrustedNonBtcRelayers"
	Msg_UpdateTrustedOracles_FullMethodName        = "/side.btcbridge.Msg/UpdateTrustedOracles"
	Msg_WithdrawToBitcoin_FullMethodName           = "/side.btcbridge.Msg/WithdrawToBitcoin"
	Msg_SubmitSignatures_FullMethodName            = "/side.btcbridge.Msg/SubmitSignatures"
	Msg_ConsolidateVaults_FullMethodName           = "/side.btcbridge.Msg/ConsolidateVaults"
	Msg_InitiateDKG_FullMethodName                 = "/side.btcbridge.Msg/InitiateDKG"
	Msg_CompleteDKG_FullMethodName                 = "/side.btcbridge.Msg/CompleteDKG"
	Msg_TransferVault_FullMethodName               = "/side.btcbridge.Msg/TransferVault"
	Msg_UpdateParams_FullMethodName                = "/side.btcbridge.Msg/UpdateParams"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	// SubmitBlockHeaders submits bitcoin block headers to the side chain.
	SubmitBlockHeaders(ctx context.Context, in *MsgSubmitBlockHeaders, opts ...grpc.CallOption) (*MsgSubmitBlockHeadersResponse, error)
	// SubmitDepositTransaction submits the bitcoin deposit transaction to the side chain.
	SubmitDepositTransaction(ctx context.Context, in *MsgSubmitDepositTransaction, opts ...grpc.CallOption) (*MsgSubmitDepositTransactionResponse, error)
	// SubmitWithdrawalTransaction submits the bitcoin withdrawal transaction to the side chain.
	SubmitWithdrawTransaction(ctx context.Context, in *MsgSubmitWithdrawTransaction, opts ...grpc.CallOption) (*MsgSubmitWithdrawTransactionResponse, error)
	// SubmitFeeRate submits the bitcoin network fee rate to the side chain.
	SubmitFeeRate(ctx context.Context, in *MsgSubmitFeeRate, opts ...grpc.CallOption) (*MsgSubmitFeeRateResponse, error)
	// UpdateTrustedNonBtcRelayers updates the trusted non-btc asset relayers.
	UpdateTrustedNonBtcRelayers(ctx context.Context, in *MsgUpdateTrustedNonBtcRelayers, opts ...grpc.CallOption) (*MsgUpdateTrustedNonBtcRelayersResponse, error)
	// UpdateTrustedOracles updates the trusted oracles.
	UpdateTrustedOracles(ctx context.Context, in *MsgUpdateTrustedOracles, opts ...grpc.CallOption) (*MsgUpdateTrustedOraclesResponse, error)
	// WithdrawToBitcoin withdraws the asset to bitcoin.
	WithdrawToBitcoin(ctx context.Context, in *MsgWithdrawToBitcoin, opts ...grpc.CallOption) (*MsgWithdrawToBitcoinResponse, error)
	// SubmitSignatures submits the signatures of the signing request to the side chain.
	SubmitSignatures(ctx context.Context, in *MsgSubmitSignatures, opts ...grpc.CallOption) (*MsgSubmitSignaturesResponse, error)
	// ConsolidateVaults performs the utxo consolidation for the given vaults.
	ConsolidateVaults(ctx context.Context, in *MsgConsolidateVaults, opts ...grpc.CallOption) (*MsgConsolidateVaultsResponse, error)
	// InitiateDKG initiates the DKG request.
	InitiateDKG(ctx context.Context, in *MsgInitiateDKG, opts ...grpc.CallOption) (*MsgInitiateDKGResponse, error)
	// CompleteDKG completes the given DKG request.
	CompleteDKG(ctx context.Context, in *MsgCompleteDKG, opts ...grpc.CallOption) (*MsgCompleteDKGResponse, error)
	// TransferVault transfers the vault asset from the source version to the destination version.
	TransferVault(ctx context.Context, in *MsgTransferVault, opts ...grpc.CallOption) (*MsgTransferVaultResponse, error)
	// UpdateParams defines a governance operation for updating the x/btcbridge module
	// parameters. The authority defaults to the x/gov module account.
	//
	// Since: cosmos-sdk 0.47
	UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) SubmitBlockHeaders(ctx context.Context, in *MsgSubmitBlockHeaders, opts ...grpc.CallOption) (*MsgSubmitBlockHeadersResponse, error) {
	out := new(MsgSubmitBlockHeadersResponse)
	err := c.cc.Invoke(ctx, Msg_SubmitBlockHeaders_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) SubmitDepositTransaction(ctx context.Context, in *MsgSubmitDepositTransaction, opts ...grpc.CallOption) (*MsgSubmitDepositTransactionResponse, error) {
	out := new(MsgSubmitDepositTransactionResponse)
	err := c.cc.Invoke(ctx, Msg_SubmitDepositTransaction_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) SubmitWithdrawTransaction(ctx context.Context, in *MsgSubmitWithdrawTransaction, opts ...grpc.CallOption) (*MsgSubmitWithdrawTransactionResponse, error) {
	out := new(MsgSubmitWithdrawTransactionResponse)
	err := c.cc.Invoke(ctx, Msg_SubmitWithdrawTransaction_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) SubmitFeeRate(ctx context.Context, in *MsgSubmitFeeRate, opts ...grpc.CallOption) (*MsgSubmitFeeRateResponse, error) {
	out := new(MsgSubmitFeeRateResponse)
	err := c.cc.Invoke(ctx, Msg_SubmitFeeRate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateTrustedNonBtcRelayers(ctx context.Context, in *MsgUpdateTrustedNonBtcRelayers, opts ...grpc.CallOption) (*MsgUpdateTrustedNonBtcRelayersResponse, error) {
	out := new(MsgUpdateTrustedNonBtcRelayersResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateTrustedNonBtcRelayers_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateTrustedOracles(ctx context.Context, in *MsgUpdateTrustedOracles, opts ...grpc.CallOption) (*MsgUpdateTrustedOraclesResponse, error) {
	out := new(MsgUpdateTrustedOraclesResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateTrustedOracles_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) WithdrawToBitcoin(ctx context.Context, in *MsgWithdrawToBitcoin, opts ...grpc.CallOption) (*MsgWithdrawToBitcoinResponse, error) {
	out := new(MsgWithdrawToBitcoinResponse)
	err := c.cc.Invoke(ctx, Msg_WithdrawToBitcoin_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) SubmitSignatures(ctx context.Context, in *MsgSubmitSignatures, opts ...grpc.CallOption) (*MsgSubmitSignaturesResponse, error) {
	out := new(MsgSubmitSignaturesResponse)
	err := c.cc.Invoke(ctx, Msg_SubmitSignatures_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ConsolidateVaults(ctx context.Context, in *MsgConsolidateVaults, opts ...grpc.CallOption) (*MsgConsolidateVaultsResponse, error) {
	out := new(MsgConsolidateVaultsResponse)
	err := c.cc.Invoke(ctx, Msg_ConsolidateVaults_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) InitiateDKG(ctx context.Context, in *MsgInitiateDKG, opts ...grpc.CallOption) (*MsgInitiateDKGResponse, error) {
	out := new(MsgInitiateDKGResponse)
	err := c.cc.Invoke(ctx, Msg_InitiateDKG_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CompleteDKG(ctx context.Context, in *MsgCompleteDKG, opts ...grpc.CallOption) (*MsgCompleteDKGResponse, error) {
	out := new(MsgCompleteDKGResponse)
	err := c.cc.Invoke(ctx, Msg_CompleteDKG_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) TransferVault(ctx context.Context, in *MsgTransferVault, opts ...grpc.CallOption) (*MsgTransferVaultResponse, error) {
	out := new(MsgTransferVaultResponse)
	err := c.cc.Invoke(ctx, Msg_TransferVault_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error) {
	out := new(MsgUpdateParamsResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateParams_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
type MsgServer interface {
	// SubmitBlockHeaders submits bitcoin block headers to the side chain.
	SubmitBlockHeaders(context.Context, *MsgSubmitBlockHeaders) (*MsgSubmitBlockHeadersResponse, error)
	// SubmitDepositTransaction submits the bitcoin deposit transaction to the side chain.
	SubmitDepositTransaction(context.Context, *MsgSubmitDepositTransaction) (*MsgSubmitDepositTransactionResponse, error)
	// SubmitWithdrawalTransaction submits the bitcoin withdrawal transaction to the side chain.
	SubmitWithdrawTransaction(context.Context, *MsgSubmitWithdrawTransaction) (*MsgSubmitWithdrawTransactionResponse, error)
	// SubmitFeeRate submits the bitcoin network fee rate to the side chain.
	SubmitFeeRate(context.Context, *MsgSubmitFeeRate) (*MsgSubmitFeeRateResponse, error)
	// UpdateTrustedNonBtcRelayers updates the trusted non-btc asset relayers.
	UpdateTrustedNonBtcRelayers(context.Context, *MsgUpdateTrustedNonBtcRelayers) (*MsgUpdateTrustedNonBtcRelayersResponse, error)
	// UpdateTrustedOracles updates the trusted oracles.
	UpdateTrustedOracles(context.Context, *MsgUpdateTrustedOracles) (*MsgUpdateTrustedOraclesResponse, error)
	// WithdrawToBitcoin withdraws the asset to bitcoin.
	WithdrawToBitcoin(context.Context, *MsgWithdrawToBitcoin) (*MsgWithdrawToBitcoinResponse, error)
	// SubmitSignatures submits the signatures of the signing request to the side chain.
	SubmitSignatures(context.Context, *MsgSubmitSignatures) (*MsgSubmitSignaturesResponse, error)
	// ConsolidateVaults performs the utxo consolidation for the given vaults.
	ConsolidateVaults(context.Context, *MsgConsolidateVaults) (*MsgConsolidateVaultsResponse, error)
	// InitiateDKG initiates the DKG request.
	InitiateDKG(context.Context, *MsgInitiateDKG) (*MsgInitiateDKGResponse, error)
	// CompleteDKG completes the given DKG request.
	CompleteDKG(context.Context, *MsgCompleteDKG) (*MsgCompleteDKGResponse, error)
	// TransferVault transfers the vault asset from the source version to the destination version.
	TransferVault(context.Context, *MsgTransferVault) (*MsgTransferVaultResponse, error)
	// UpdateParams defines a governance operation for updating the x/btcbridge module
	// parameters. The authority defaults to the x/gov module account.
	//
	// Since: cosmos-sdk 0.47
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) SubmitBlockHeaders(context.Context, *MsgSubmitBlockHeaders) (*MsgSubmitBlockHeadersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitBlockHeaders not implemented")
}
func (UnimplementedMsgServer) SubmitDepositTransaction(context.Context, *MsgSubmitDepositTransaction) (*MsgSubmitDepositTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitDepositTransaction not implemented")
}
func (UnimplementedMsgServer) SubmitWithdrawTransaction(context.Context, *MsgSubmitWithdrawTransaction) (*MsgSubmitWithdrawTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitWithdrawTransaction not implemented")
}
func (UnimplementedMsgServer) SubmitFeeRate(context.Context, *MsgSubmitFeeRate) (*MsgSubmitFeeRateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitFeeRate not implemented")
}
func (UnimplementedMsgServer) UpdateTrustedNonBtcRelayers(context.Context, *MsgUpdateTrustedNonBtcRelayers) (*MsgUpdateTrustedNonBtcRelayersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTrustedNonBtcRelayers not implemented")
}
func (UnimplementedMsgServer) UpdateTrustedOracles(context.Context, *MsgUpdateTrustedOracles) (*MsgUpdateTrustedOraclesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTrustedOracles not implemented")
}
func (UnimplementedMsgServer) WithdrawToBitcoin(context.Context, *MsgWithdrawToBitcoin) (*MsgWithdrawToBitcoinResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WithdrawToBitcoin not implemented")
}
func (UnimplementedMsgServer) SubmitSignatures(context.Context, *MsgSubmitSignatures) (*MsgSubmitSignaturesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitSignatures not implemented")
}
func (UnimplementedMsgServer) ConsolidateVaults(context.Context, *MsgConsolidateVaults) (*MsgConsolidateVaultsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConsolidateVaults not implemented")
}
func (UnimplementedMsgServer) InitiateDKG(context.Context, *MsgInitiateDKG) (*MsgInitiateDKGResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitiateDKG not implemented")
}
func (UnimplementedMsgServer) CompleteDKG(context.Context, *MsgCompleteDKG) (*MsgCompleteDKGResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CompleteDKG not implemented")
}
func (UnimplementedMsgServer) TransferVault(context.Context, *MsgTransferVault) (*MsgTransferVaultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TransferVault not implemented")
}
func (UnimplementedMsgServer) UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateParams not implemented")
}
func (UnimplementedMsgServer) mustEmbedUnimplementedMsgServer() {}

// UnsafeMsgServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MsgServer will
// result in compilation errors.
type UnsafeMsgServer interface {
	mustEmbedUnimplementedMsgServer()
}

func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&Msg_ServiceDesc, srv)
}

func _Msg_SubmitBlockHeaders_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitBlockHeaders)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitBlockHeaders(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SubmitBlockHeaders_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitBlockHeaders(ctx, req.(*MsgSubmitBlockHeaders))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SubmitDepositTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitDepositTransaction)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitDepositTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SubmitDepositTransaction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitDepositTransaction(ctx, req.(*MsgSubmitDepositTransaction))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SubmitWithdrawTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitWithdrawTransaction)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitWithdrawTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SubmitWithdrawTransaction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitWithdrawTransaction(ctx, req.(*MsgSubmitWithdrawTransaction))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SubmitFeeRate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitFeeRate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitFeeRate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SubmitFeeRate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitFeeRate(ctx, req.(*MsgSubmitFeeRate))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateTrustedNonBtcRelayers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateTrustedNonBtcRelayers)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateTrustedNonBtcRelayers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateTrustedNonBtcRelayers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateTrustedNonBtcRelayers(ctx, req.(*MsgUpdateTrustedNonBtcRelayers))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateTrustedOracles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateTrustedOracles)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateTrustedOracles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateTrustedOracles_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateTrustedOracles(ctx, req.(*MsgUpdateTrustedOracles))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_WithdrawToBitcoin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgWithdrawToBitcoin)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).WithdrawToBitcoin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_WithdrawToBitcoin_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).WithdrawToBitcoin(ctx, req.(*MsgWithdrawToBitcoin))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SubmitSignatures_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitSignatures)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitSignatures(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SubmitSignatures_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitSignatures(ctx, req.(*MsgSubmitSignatures))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ConsolidateVaults_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgConsolidateVaults)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ConsolidateVaults(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_ConsolidateVaults_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ConsolidateVaults(ctx, req.(*MsgConsolidateVaults))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_InitiateDKG_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgInitiateDKG)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).InitiateDKG(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_InitiateDKG_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).InitiateDKG(ctx, req.(*MsgInitiateDKG))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CompleteDKG_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCompleteDKG)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CompleteDKG(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_CompleteDKG_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CompleteDKG(ctx, req.(*MsgCompleteDKG))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_TransferVault_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgTransferVault)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).TransferVault(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_TransferVault_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).TransferVault(ctx, req.(*MsgTransferVault))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateParams_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateParams(ctx, req.(*MsgUpdateParams))
	}
	return interceptor(ctx, in, info, handler)
}

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "side.btcbridge.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitBlockHeaders",
			Handler:    _Msg_SubmitBlockHeaders_Handler,
		},
		{
			MethodName: "SubmitDepositTransaction",
			Handler:    _Msg_SubmitDepositTransaction_Handler,
		},
		{
			MethodName: "SubmitWithdrawTransaction",
			Handler:    _Msg_SubmitWithdrawTransaction_Handler,
		},
		{
			MethodName: "SubmitFeeRate",
			Handler:    _Msg_SubmitFeeRate_Handler,
		},
		{
			MethodName: "UpdateTrustedNonBtcRelayers",
			Handler:    _Msg_UpdateTrustedNonBtcRelayers_Handler,
		},
		{
			MethodName: "UpdateTrustedOracles",
			Handler:    _Msg_UpdateTrustedOracles_Handler,
		},
		{
			MethodName: "WithdrawToBitcoin",
			Handler:    _Msg_WithdrawToBitcoin_Handler,
		},
		{
			MethodName: "SubmitSignatures",
			Handler:    _Msg_SubmitSignatures_Handler,
		},
		{
			MethodName: "ConsolidateVaults",
			Handler:    _Msg_ConsolidateVaults_Handler,
		},
		{
			MethodName: "InitiateDKG",
			Handler:    _Msg_InitiateDKG_Handler,
		},
		{
			MethodName: "CompleteDKG",
			Handler:    _Msg_CompleteDKG_Handler,
		},
		{
			MethodName: "TransferVault",
			Handler:    _Msg_TransferVault_Handler,
		},
		{
			MethodName: "UpdateParams",
			Handler:    _Msg_UpdateParams_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "side/btcbridge/tx.proto",
}
