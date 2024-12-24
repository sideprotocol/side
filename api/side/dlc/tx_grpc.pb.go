// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: side/dlc/tx.proto

package dlc

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
	Msg_SubmitNonce_FullMethodName        = "/side.dlc.Msg/SubmitNonce"
	Msg_SubmitAttestation_FullMethodName  = "/side.dlc.Msg/SubmitAttestation"
	Msg_SubmitOraclePubKey_FullMethodName = "/side.dlc.Msg/SubmitOraclePubKey"
	Msg_SubmitAgencyPubKey_FullMethodName = "/side.dlc.Msg/SubmitAgencyPubKey"
	Msg_CreateOracle_FullMethodName       = "/side.dlc.Msg/CreateOracle"
	Msg_CreateAgency_FullMethodName       = "/side.dlc.Msg/CreateAgency"
	Msg_UpdateParams_FullMethodName       = "/side.dlc.Msg/UpdateParams"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	SubmitNonce(ctx context.Context, in *MsgSubmitNonce, opts ...grpc.CallOption) (*MsgSubmitNonceResponse, error)
	SubmitAttestation(ctx context.Context, in *MsgSubmitAttestation, opts ...grpc.CallOption) (*MsgSubmitAttestationResponse, error)
	SubmitOraclePubKey(ctx context.Context, in *MsgSubmitOraclePubKey, opts ...grpc.CallOption) (*MsgSubmitOraclePubKeyResponse, error)
	SubmitAgencyPubKey(ctx context.Context, in *MsgSubmitAgencyPubKey, opts ...grpc.CallOption) (*MsgSubmitAgencyPubKeyResponse, error)
	CreateOracle(ctx context.Context, in *MsgCreateOracle, opts ...grpc.CallOption) (*MsgCreateOracleResponse, error)
	CreateAgency(ctx context.Context, in *MsgCreateAgency, opts ...grpc.CallOption) (*MsgCreateAgencyResponse, error)
	// UpdateParams defines a governance operation for updating the x/dlc module
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

func (c *msgClient) SubmitNonce(ctx context.Context, in *MsgSubmitNonce, opts ...grpc.CallOption) (*MsgSubmitNonceResponse, error) {
	out := new(MsgSubmitNonceResponse)
	err := c.cc.Invoke(ctx, Msg_SubmitNonce_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) SubmitAttestation(ctx context.Context, in *MsgSubmitAttestation, opts ...grpc.CallOption) (*MsgSubmitAttestationResponse, error) {
	out := new(MsgSubmitAttestationResponse)
	err := c.cc.Invoke(ctx, Msg_SubmitAttestation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) SubmitOraclePubKey(ctx context.Context, in *MsgSubmitOraclePubKey, opts ...grpc.CallOption) (*MsgSubmitOraclePubKeyResponse, error) {
	out := new(MsgSubmitOraclePubKeyResponse)
	err := c.cc.Invoke(ctx, Msg_SubmitOraclePubKey_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) SubmitAgencyPubKey(ctx context.Context, in *MsgSubmitAgencyPubKey, opts ...grpc.CallOption) (*MsgSubmitAgencyPubKeyResponse, error) {
	out := new(MsgSubmitAgencyPubKeyResponse)
	err := c.cc.Invoke(ctx, Msg_SubmitAgencyPubKey_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CreateOracle(ctx context.Context, in *MsgCreateOracle, opts ...grpc.CallOption) (*MsgCreateOracleResponse, error) {
	out := new(MsgCreateOracleResponse)
	err := c.cc.Invoke(ctx, Msg_CreateOracle_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CreateAgency(ctx context.Context, in *MsgCreateAgency, opts ...grpc.CallOption) (*MsgCreateAgencyResponse, error) {
	out := new(MsgCreateAgencyResponse)
	err := c.cc.Invoke(ctx, Msg_CreateAgency_FullMethodName, in, out, opts...)
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
	SubmitNonce(context.Context, *MsgSubmitNonce) (*MsgSubmitNonceResponse, error)
	SubmitAttestation(context.Context, *MsgSubmitAttestation) (*MsgSubmitAttestationResponse, error)
	SubmitOraclePubKey(context.Context, *MsgSubmitOraclePubKey) (*MsgSubmitOraclePubKeyResponse, error)
	SubmitAgencyPubKey(context.Context, *MsgSubmitAgencyPubKey) (*MsgSubmitAgencyPubKeyResponse, error)
	CreateOracle(context.Context, *MsgCreateOracle) (*MsgCreateOracleResponse, error)
	CreateAgency(context.Context, *MsgCreateAgency) (*MsgCreateAgencyResponse, error)
	// UpdateParams defines a governance operation for updating the x/dlc module
	// parameters. The authority defaults to the x/gov module account.
	//
	// Since: cosmos-sdk 0.47
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) SubmitNonce(context.Context, *MsgSubmitNonce) (*MsgSubmitNonceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitNonce not implemented")
}
func (UnimplementedMsgServer) SubmitAttestation(context.Context, *MsgSubmitAttestation) (*MsgSubmitAttestationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitAttestation not implemented")
}
func (UnimplementedMsgServer) SubmitOraclePubKey(context.Context, *MsgSubmitOraclePubKey) (*MsgSubmitOraclePubKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitOraclePubKey not implemented")
}
func (UnimplementedMsgServer) SubmitAgencyPubKey(context.Context, *MsgSubmitAgencyPubKey) (*MsgSubmitAgencyPubKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitAgencyPubKey not implemented")
}
func (UnimplementedMsgServer) CreateOracle(context.Context, *MsgCreateOracle) (*MsgCreateOracleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOracle not implemented")
}
func (UnimplementedMsgServer) CreateAgency(context.Context, *MsgCreateAgency) (*MsgCreateAgencyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAgency not implemented")
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

func _Msg_SubmitNonce_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitNonce)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitNonce(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SubmitNonce_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitNonce(ctx, req.(*MsgSubmitNonce))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SubmitAttestation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitAttestation)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitAttestation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SubmitAttestation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitAttestation(ctx, req.(*MsgSubmitAttestation))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SubmitOraclePubKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitOraclePubKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitOraclePubKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SubmitOraclePubKey_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitOraclePubKey(ctx, req.(*MsgSubmitOraclePubKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SubmitAgencyPubKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitAgencyPubKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitAgencyPubKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SubmitAgencyPubKey_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitAgencyPubKey(ctx, req.(*MsgSubmitAgencyPubKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CreateOracle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateOracle)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateOracle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_CreateOracle_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateOracle(ctx, req.(*MsgCreateOracle))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CreateAgency_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateAgency)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateAgency(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_CreateAgency_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateAgency(ctx, req.(*MsgCreateAgency))
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
	ServiceName: "side.dlc.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitNonce",
			Handler:    _Msg_SubmitNonce_Handler,
		},
		{
			MethodName: "SubmitAttestation",
			Handler:    _Msg_SubmitAttestation_Handler,
		},
		{
			MethodName: "SubmitOraclePubKey",
			Handler:    _Msg_SubmitOraclePubKey_Handler,
		},
		{
			MethodName: "SubmitAgencyPubKey",
			Handler:    _Msg_SubmitAgencyPubKey_Handler,
		},
		{
			MethodName: "CreateOracle",
			Handler:    _Msg_CreateOracle_Handler,
		},
		{
			MethodName: "CreateAgency",
			Handler:    _Msg_CreateAgency_Handler,
		},
		{
			MethodName: "UpdateParams",
			Handler:    _Msg_UpdateParams_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "side/dlc/tx.proto",
}
