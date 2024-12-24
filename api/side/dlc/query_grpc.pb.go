// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: side/dlc/query.proto

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
	Query_Params_FullMethodName       = "/side.dlc.Query/Params"
	Query_Events_FullMethodName       = "/side.dlc.Query/Events"
	Query_Attestations_FullMethodName = "/side.dlc.Query/Attestations"
	Query_Price_FullMethodName        = "/side.dlc.Query/Price"
	Query_Nonces_FullMethodName       = "/side.dlc.Query/Nonces"
	Query_CountNonces_FullMethodName  = "/side.dlc.Query/CountNonces"
	Query_Oracles_FullMethodName      = "/side.dlc.Query/Oracles"
	Query_Agencies_FullMethodName     = "/side.dlc.Query/Agencies"
)

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type QueryClient interface {
	// Params queries the parameters of the module.
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
	// Announcements queries the announcements by the given status.
	Events(ctx context.Context, in *QueryPriceEventRequest, opts ...grpc.CallOption) (*QueryPriceEventResponse, error)
	Attestations(ctx context.Context, in *QueryAttestationRequest, opts ...grpc.CallOption) (*QueryAttestationResponse, error)
	// Price queries the current price by the given symbol.
	Price(ctx context.Context, in *QueryPriceRequest, opts ...grpc.CallOption) (*QueryPriceResponse, error)
	Nonces(ctx context.Context, in *QueryNoncesRequest, opts ...grpc.CallOption) (*QueryNoncesResponse, error)
	CountNonces(ctx context.Context, in *QueryCountNoncesRequest, opts ...grpc.CallOption) (*QueryCountNoncesResponse, error)
	Oracles(ctx context.Context, in *QueryOraclesRequest, opts ...grpc.CallOption) (*QueryOraclesResponse, error)
	Agencies(ctx context.Context, in *QueryAgenciesRequest, opts ...grpc.CallOption) (*QueryAgenciesResponse, error)
}

type queryClient struct {
	cc grpc.ClientConnInterface
}

func NewQueryClient(cc grpc.ClientConnInterface) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, Query_Params_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Events(ctx context.Context, in *QueryPriceEventRequest, opts ...grpc.CallOption) (*QueryPriceEventResponse, error) {
	out := new(QueryPriceEventResponse)
	err := c.cc.Invoke(ctx, Query_Events_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Attestations(ctx context.Context, in *QueryAttestationRequest, opts ...grpc.CallOption) (*QueryAttestationResponse, error) {
	out := new(QueryAttestationResponse)
	err := c.cc.Invoke(ctx, Query_Attestations_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Price(ctx context.Context, in *QueryPriceRequest, opts ...grpc.CallOption) (*QueryPriceResponse, error) {
	out := new(QueryPriceResponse)
	err := c.cc.Invoke(ctx, Query_Price_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Nonces(ctx context.Context, in *QueryNoncesRequest, opts ...grpc.CallOption) (*QueryNoncesResponse, error) {
	out := new(QueryNoncesResponse)
	err := c.cc.Invoke(ctx, Query_Nonces_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) CountNonces(ctx context.Context, in *QueryCountNoncesRequest, opts ...grpc.CallOption) (*QueryCountNoncesResponse, error) {
	out := new(QueryCountNoncesResponse)
	err := c.cc.Invoke(ctx, Query_CountNonces_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Oracles(ctx context.Context, in *QueryOraclesRequest, opts ...grpc.CallOption) (*QueryOraclesResponse, error) {
	out := new(QueryOraclesResponse)
	err := c.cc.Invoke(ctx, Query_Oracles_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Agencies(ctx context.Context, in *QueryAgenciesRequest, opts ...grpc.CallOption) (*QueryAgenciesResponse, error) {
	out := new(QueryAgenciesResponse)
	err := c.cc.Invoke(ctx, Query_Agencies_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
// All implementations must embed UnimplementedQueryServer
// for forward compatibility
type QueryServer interface {
	// Params queries the parameters of the module.
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	// Announcements queries the announcements by the given status.
	Events(context.Context, *QueryPriceEventRequest) (*QueryPriceEventResponse, error)
	Attestations(context.Context, *QueryAttestationRequest) (*QueryAttestationResponse, error)
	// Price queries the current price by the given symbol.
	Price(context.Context, *QueryPriceRequest) (*QueryPriceResponse, error)
	Nonces(context.Context, *QueryNoncesRequest) (*QueryNoncesResponse, error)
	CountNonces(context.Context, *QueryCountNoncesRequest) (*QueryCountNoncesResponse, error)
	Oracles(context.Context, *QueryOraclesRequest) (*QueryOraclesResponse, error)
	Agencies(context.Context, *QueryAgenciesRequest) (*QueryAgenciesResponse, error)
	mustEmbedUnimplementedQueryServer()
}

// UnimplementedQueryServer must be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (UnimplementedQueryServer) Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (UnimplementedQueryServer) Events(context.Context, *QueryPriceEventRequest) (*QueryPriceEventResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Events not implemented")
}
func (UnimplementedQueryServer) Attestations(context.Context, *QueryAttestationRequest) (*QueryAttestationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Attestations not implemented")
}
func (UnimplementedQueryServer) Price(context.Context, *QueryPriceRequest) (*QueryPriceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Price not implemented")
}
func (UnimplementedQueryServer) Nonces(context.Context, *QueryNoncesRequest) (*QueryNoncesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Nonces not implemented")
}
func (UnimplementedQueryServer) CountNonces(context.Context, *QueryCountNoncesRequest) (*QueryCountNoncesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CountNonces not implemented")
}
func (UnimplementedQueryServer) Oracles(context.Context, *QueryOraclesRequest) (*QueryOraclesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Oracles not implemented")
}
func (UnimplementedQueryServer) Agencies(context.Context, *QueryAgenciesRequest) (*QueryAgenciesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Agencies not implemented")
}
func (UnimplementedQueryServer) mustEmbedUnimplementedQueryServer() {}

// UnsafeQueryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QueryServer will
// result in compilation errors.
type UnsafeQueryServer interface {
	mustEmbedUnimplementedQueryServer()
}

func RegisterQueryServer(s grpc.ServiceRegistrar, srv QueryServer) {
	s.RegisterService(&Query_ServiceDesc, srv)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_Params_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Events_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryPriceEventRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Events(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_Events_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Events(ctx, req.(*QueryPriceEventRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Attestations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAttestationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Attestations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_Attestations_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Attestations(ctx, req.(*QueryAttestationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Price_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryPriceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Price(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_Price_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Price(ctx, req.(*QueryPriceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Nonces_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryNoncesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Nonces(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_Nonces_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Nonces(ctx, req.(*QueryNoncesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_CountNonces_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryCountNoncesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).CountNonces(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_CountNonces_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).CountNonces(ctx, req.(*QueryCountNoncesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Oracles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryOraclesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Oracles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_Oracles_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Oracles(ctx, req.(*QueryOraclesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Agencies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAgenciesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Agencies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_Agencies_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Agencies(ctx, req.(*QueryAgenciesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Query_ServiceDesc is the grpc.ServiceDesc for Query service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Query_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "side.dlc.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Params",
			Handler:    _Query_Params_Handler,
		},
		{
			MethodName: "Events",
			Handler:    _Query_Events_Handler,
		},
		{
			MethodName: "Attestations",
			Handler:    _Query_Attestations_Handler,
		},
		{
			MethodName: "Price",
			Handler:    _Query_Price_Handler,
		},
		{
			MethodName: "Nonces",
			Handler:    _Query_Nonces_Handler,
		},
		{
			MethodName: "CountNonces",
			Handler:    _Query_CountNonces_Handler,
		},
		{
			MethodName: "Oracles",
			Handler:    _Query_Oracles_Handler,
		},
		{
			MethodName: "Agencies",
			Handler:    _Query_Agencies_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "side/dlc/query.proto",
}
