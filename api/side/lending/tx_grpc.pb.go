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
	Msg_CreatePool_FullMethodName      = "/side.lending.Msg/CreatePool"
	Msg_AddLiquidity_FullMethodName    = "/side.lending.Msg/AddLiquidity"
	Msg_RemoveLiquidity_FullMethodName = "/side.lending.Msg/RemoveLiquidity"
	Msg_Apply_FullMethodName           = "/side.lending.Msg/Apply"
	Msg_Approve_FullMethodName         = "/side.lending.Msg/Approve"
	Msg_Redeem_FullMethodName          = "/side.lending.Msg/Redeem"
	Msg_Repay_FullMethodName           = "/side.lending.Msg/Repay"
	Msg_Close_FullMethodName           = "/side.lending.Msg/Close"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	CreatePool(ctx context.Context, in *MsgCreatePool, opts ...grpc.CallOption) (*MsgCreatePoolResponse, error)
	AddLiquidity(ctx context.Context, in *MsgAddLiquidity, opts ...grpc.CallOption) (*MsgAddLiquidityResponse, error)
	RemoveLiquidity(ctx context.Context, in *MsgRemoveLiquidity, opts ...grpc.CallOption) (*MsgRemoveLiquidityResponse, error)
	Apply(ctx context.Context, in *MsgApply, opts ...grpc.CallOption) (*MsgApplyResponse, error)
	Approve(ctx context.Context, in *MsgApprove, opts ...grpc.CallOption) (*MsgApproveResponse, error)
	Redeem(ctx context.Context, in *MsgRedeem, opts ...grpc.CallOption) (*MsgRedeemResponse, error)
	Repay(ctx context.Context, in *MsgRepay, opts ...grpc.CallOption) (*MsgRepayResponse, error)
	Close(ctx context.Context, in *MsgClose, opts ...grpc.CallOption) (*MsgCloseResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) CreatePool(ctx context.Context, in *MsgCreatePool, opts ...grpc.CallOption) (*MsgCreatePoolResponse, error) {
	out := new(MsgCreatePoolResponse)
	err := c.cc.Invoke(ctx, Msg_CreatePool_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
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

func (c *msgClient) Apply(ctx context.Context, in *MsgApply, opts ...grpc.CallOption) (*MsgApplyResponse, error) {
	out := new(MsgApplyResponse)
	err := c.cc.Invoke(ctx, Msg_Apply_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) Approve(ctx context.Context, in *MsgApprove, opts ...grpc.CallOption) (*MsgApproveResponse, error) {
	out := new(MsgApproveResponse)
	err := c.cc.Invoke(ctx, Msg_Approve_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) Redeem(ctx context.Context, in *MsgRedeem, opts ...grpc.CallOption) (*MsgRedeemResponse, error) {
	out := new(MsgRedeemResponse)
	err := c.cc.Invoke(ctx, Msg_Redeem_FullMethodName, in, out, opts...)
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

func (c *msgClient) Close(ctx context.Context, in *MsgClose, opts ...grpc.CallOption) (*MsgCloseResponse, error) {
	out := new(MsgCloseResponse)
	err := c.cc.Invoke(ctx, Msg_Close_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
type MsgServer interface {
	CreatePool(context.Context, *MsgCreatePool) (*MsgCreatePoolResponse, error)
	AddLiquidity(context.Context, *MsgAddLiquidity) (*MsgAddLiquidityResponse, error)
	RemoveLiquidity(context.Context, *MsgRemoveLiquidity) (*MsgRemoveLiquidityResponse, error)
	Apply(context.Context, *MsgApply) (*MsgApplyResponse, error)
	Approve(context.Context, *MsgApprove) (*MsgApproveResponse, error)
	Redeem(context.Context, *MsgRedeem) (*MsgRedeemResponse, error)
	Repay(context.Context, *MsgRepay) (*MsgRepayResponse, error)
	Close(context.Context, *MsgClose) (*MsgCloseResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) CreatePool(context.Context, *MsgCreatePool) (*MsgCreatePoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePool not implemented")
}
func (UnimplementedMsgServer) AddLiquidity(context.Context, *MsgAddLiquidity) (*MsgAddLiquidityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddLiquidity not implemented")
}
func (UnimplementedMsgServer) RemoveLiquidity(context.Context, *MsgRemoveLiquidity) (*MsgRemoveLiquidityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveLiquidity not implemented")
}
func (UnimplementedMsgServer) Apply(context.Context, *MsgApply) (*MsgApplyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Apply not implemented")
}
func (UnimplementedMsgServer) Approve(context.Context, *MsgApprove) (*MsgApproveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Approve not implemented")
}
func (UnimplementedMsgServer) Redeem(context.Context, *MsgRedeem) (*MsgRedeemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Redeem not implemented")
}
func (UnimplementedMsgServer) Repay(context.Context, *MsgRepay) (*MsgRepayResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Repay not implemented")
}
func (UnimplementedMsgServer) Close(context.Context, *MsgClose) (*MsgCloseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
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

func _Msg_CreatePool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreatePool)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreatePool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_CreatePool_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreatePool(ctx, req.(*MsgCreatePool))
	}
	return interceptor(ctx, in, info, handler)
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

func _Msg_Apply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgApply)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Apply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_Apply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Apply(ctx, req.(*MsgApply))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Approve_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgApprove)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Approve(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_Approve_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Approve(ctx, req.(*MsgApprove))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Redeem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRedeem)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Redeem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_Redeem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Redeem(ctx, req.(*MsgRedeem))
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

func _Msg_Close_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgClose)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Close(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_Close_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Close(ctx, req.(*MsgClose))
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
			MethodName: "CreatePool",
			Handler:    _Msg_CreatePool_Handler,
		},
		{
			MethodName: "AddLiquidity",
			Handler:    _Msg_AddLiquidity_Handler,
		},
		{
			MethodName: "RemoveLiquidity",
			Handler:    _Msg_RemoveLiquidity_Handler,
		},
		{
			MethodName: "Apply",
			Handler:    _Msg_Apply_Handler,
		},
		{
			MethodName: "Approve",
			Handler:    _Msg_Approve_Handler,
		},
		{
			MethodName: "Redeem",
			Handler:    _Msg_Redeem_Handler,
		},
		{
			MethodName: "Repay",
			Handler:    _Msg_Repay_Handler,
		},
		{
			MethodName: "Close",
			Handler:    _Msg_Close_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "side/lending/tx.proto",
}
