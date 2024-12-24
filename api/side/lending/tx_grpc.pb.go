// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: side/lending/tx.proto

package lending

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
	Msg_AddLiquidity_FullMethodName        = "/side.lending.Msg/AddLiquidity"
	Msg_RemoveLiquidity_FullMethodName     = "/side.lending.Msg/RemoveLiquidity"
	Msg_RequestVaultAddress_FullMethodName = "/side.lending.Msg/RequestVaultAddress"
	Msg_SubmitFundingTx_FullMethodName     = "/side.lending.Msg/SubmitFundingTx"
	Msg_CreateLoan_FullMethodName          = "/side.lending.Msg/CreateLoan"
	Msg_Repay_FullMethodName               = "/side.lending.Msg/Repay"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	AddLiquidity(ctx context.Context, in *MsgAddLiquidity, opts ...grpc.CallOption) (*MsgAddLiquidityResponse, error)
	RemoveLiquidity(ctx context.Context, in *MsgRemoveLiquidity, opts ...grpc.CallOption) (*MsgRemoveLiquidityResponse, error)
	RequestVaultAddress(ctx context.Context, in *MsgRequestVaultAddress, opts ...grpc.CallOption) (*MsgRequestVaultAddressResponse, error)
	SubmitFundingTx(ctx context.Context, in *MsgSubmitFundingTx, opts ...grpc.CallOption) (*MsgSubmitFundingTxResponse, error)
	CreateLoan(ctx context.Context, in *MsgCreateLoan, opts ...grpc.CallOption) (*MsgCreateLoanResponse, error)
	Repay(ctx context.Context, in *MsgRepay, opts ...grpc.CallOption) (*MsgRepayResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) AddLiquidity(ctx context.Context, in *MsgAddLiquidity, opts ...grpc.CallOption) (*MsgAddLiquidityResponse, error) {
	out := new(MsgAddLiquidityResponse)
	err := c.cc.Invoke(ctx, Msg_AddLiquidity_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RemoveLiquidity(ctx context.Context, in *MsgRemoveLiquidity, opts ...grpc.CallOption) (*MsgRemoveLiquidityResponse, error) {
	out := new(MsgRemoveLiquidityResponse)
	err := c.cc.Invoke(ctx, Msg_RemoveLiquidity_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RequestVaultAddress(ctx context.Context, in *MsgRequestVaultAddress, opts ...grpc.CallOption) (*MsgRequestVaultAddressResponse, error) {
	out := new(MsgRequestVaultAddressResponse)
	err := c.cc.Invoke(ctx, Msg_RequestVaultAddress_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) SubmitFundingTx(ctx context.Context, in *MsgSubmitFundingTx, opts ...grpc.CallOption) (*MsgSubmitFundingTxResponse, error) {
	out := new(MsgSubmitFundingTxResponse)
	err := c.cc.Invoke(ctx, Msg_SubmitFundingTx_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CreateLoan(ctx context.Context, in *MsgCreateLoan, opts ...grpc.CallOption) (*MsgCreateLoanResponse, error) {
	out := new(MsgCreateLoanResponse)
	err := c.cc.Invoke(ctx, Msg_CreateLoan_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) Repay(ctx context.Context, in *MsgRepay, opts ...grpc.CallOption) (*MsgRepayResponse, error) {
	out := new(MsgRepayResponse)
	err := c.cc.Invoke(ctx, Msg_Repay_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
type MsgServer interface {
	AddLiquidity(context.Context, *MsgAddLiquidity) (*MsgAddLiquidityResponse, error)
	RemoveLiquidity(context.Context, *MsgRemoveLiquidity) (*MsgRemoveLiquidityResponse, error)
	RequestVaultAddress(context.Context, *MsgRequestVaultAddress) (*MsgRequestVaultAddressResponse, error)
	SubmitFundingTx(context.Context, *MsgSubmitFundingTx) (*MsgSubmitFundingTxResponse, error)
	CreateLoan(context.Context, *MsgCreateLoan) (*MsgCreateLoanResponse, error)
	Repay(context.Context, *MsgRepay) (*MsgRepayResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) AddLiquidity(context.Context, *MsgAddLiquidity) (*MsgAddLiquidityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddLiquidity not implemented")
}
func (UnimplementedMsgServer) RemoveLiquidity(context.Context, *MsgRemoveLiquidity) (*MsgRemoveLiquidityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveLiquidity not implemented")
}
func (UnimplementedMsgServer) RequestVaultAddress(context.Context, *MsgRequestVaultAddress) (*MsgRequestVaultAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestVaultAddress not implemented")
}
func (UnimplementedMsgServer) SubmitFundingTx(context.Context, *MsgSubmitFundingTx) (*MsgSubmitFundingTxResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitFundingTx not implemented")
}
func (UnimplementedMsgServer) CreateLoan(context.Context, *MsgCreateLoan) (*MsgCreateLoanResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateLoan not implemented")
}
func (UnimplementedMsgServer) Repay(context.Context, *MsgRepay) (*MsgRepayResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Repay not implemented")
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

func _Msg_AddLiquidity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAddLiquidity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).AddLiquidity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_AddLiquidity_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).AddLiquidity(ctx, req.(*MsgAddLiquidity))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RemoveLiquidity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRemoveLiquidity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RemoveLiquidity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_RemoveLiquidity_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RemoveLiquidity(ctx, req.(*MsgRemoveLiquidity))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RequestVaultAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRequestVaultAddress)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RequestVaultAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_RequestVaultAddress_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RequestVaultAddress(ctx, req.(*MsgRequestVaultAddress))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SubmitFundingTx_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitFundingTx)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitFundingTx(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SubmitFundingTx_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitFundingTx(ctx, req.(*MsgSubmitFundingTx))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CreateLoan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateLoan)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateLoan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_CreateLoan_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateLoan(ctx, req.(*MsgCreateLoan))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Repay_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRepay)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Repay(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_Repay_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Repay(ctx, req.(*MsgRepay))
	}
	return interceptor(ctx, in, info, handler)
}

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "side.lending.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddLiquidity",
			Handler:    _Msg_AddLiquidity_Handler,
		},
		{
			MethodName: "RemoveLiquidity",
			Handler:    _Msg_RemoveLiquidity_Handler,
		},
		{
			MethodName: "RequestVaultAddress",
			Handler:    _Msg_RequestVaultAddress_Handler,
		},
		{
			MethodName: "SubmitFundingTx",
			Handler:    _Msg_SubmitFundingTx_Handler,
		},
		{
			MethodName: "CreateLoan",
			Handler:    _Msg_CreateLoan_Handler,
		},
		{
			MethodName: "Repay",
			Handler:    _Msg_Repay_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "side/lending/tx.proto",
}
