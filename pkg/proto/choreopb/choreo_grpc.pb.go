// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.24.3
// source: choreo.proto

package choreopb

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

// ChoreoClient is the client API for Choreo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChoreoClient interface {
	Get(ctx context.Context, in *Get_Request, opts ...grpc.CallOption) (*Get_Response, error)
	Apply(ctx context.Context, in *Apply_Request, opts ...grpc.CallOption) (*Apply_Response, error)
	// rpc Watch (Watch.Request) returns (stream Watch.Response) {}
	Start(ctx context.Context, in *Start_Request, opts ...grpc.CallOption) (*Start_Response, error)
	Stop(ctx context.Context, in *Stop_Request, opts ...grpc.CallOption) (*Stop_Response, error)
	Once(ctx context.Context, in *Once_Request, opts ...grpc.CallOption) (*Once_Response, error)
	Load(ctx context.Context, in *Load_Request, opts ...grpc.CallOption) (*Load_Response, error)
}

type choreoClient struct {
	cc grpc.ClientConnInterface
}

func NewChoreoClient(cc grpc.ClientConnInterface) ChoreoClient {
	return &choreoClient{cc}
}

func (c *choreoClient) Get(ctx context.Context, in *Get_Request, opts ...grpc.CallOption) (*Get_Response, error) {
	out := new(Get_Response)
	err := c.cc.Invoke(ctx, "/choreopb.Choreo/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *choreoClient) Apply(ctx context.Context, in *Apply_Request, opts ...grpc.CallOption) (*Apply_Response, error) {
	out := new(Apply_Response)
	err := c.cc.Invoke(ctx, "/choreopb.Choreo/Apply", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *choreoClient) Start(ctx context.Context, in *Start_Request, opts ...grpc.CallOption) (*Start_Response, error) {
	out := new(Start_Response)
	err := c.cc.Invoke(ctx, "/choreopb.Choreo/Start", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *choreoClient) Stop(ctx context.Context, in *Stop_Request, opts ...grpc.CallOption) (*Stop_Response, error) {
	out := new(Stop_Response)
	err := c.cc.Invoke(ctx, "/choreopb.Choreo/Stop", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *choreoClient) Once(ctx context.Context, in *Once_Request, opts ...grpc.CallOption) (*Once_Response, error) {
	out := new(Once_Response)
	err := c.cc.Invoke(ctx, "/choreopb.Choreo/Once", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *choreoClient) Load(ctx context.Context, in *Load_Request, opts ...grpc.CallOption) (*Load_Response, error) {
	out := new(Load_Response)
	err := c.cc.Invoke(ctx, "/choreopb.Choreo/Load", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChoreoServer is the server API for Choreo service.
// All implementations must embed UnimplementedChoreoServer
// for forward compatibility
type ChoreoServer interface {
	Get(context.Context, *Get_Request) (*Get_Response, error)
	Apply(context.Context, *Apply_Request) (*Apply_Response, error)
	// rpc Watch (Watch.Request) returns (stream Watch.Response) {}
	Start(context.Context, *Start_Request) (*Start_Response, error)
	Stop(context.Context, *Stop_Request) (*Stop_Response, error)
	Once(context.Context, *Once_Request) (*Once_Response, error)
	Load(context.Context, *Load_Request) (*Load_Response, error)
	mustEmbedUnimplementedChoreoServer()
}

// UnimplementedChoreoServer must be embedded to have forward compatible implementations.
type UnimplementedChoreoServer struct {
}

func (UnimplementedChoreoServer) Get(context.Context, *Get_Request) (*Get_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedChoreoServer) Apply(context.Context, *Apply_Request) (*Apply_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Apply not implemented")
}
func (UnimplementedChoreoServer) Start(context.Context, *Start_Request) (*Start_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Start not implemented")
}
func (UnimplementedChoreoServer) Stop(context.Context, *Stop_Request) (*Stop_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
func (UnimplementedChoreoServer) Once(context.Context, *Once_Request) (*Once_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Once not implemented")
}
func (UnimplementedChoreoServer) Load(context.Context, *Load_Request) (*Load_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Load not implemented")
}
func (UnimplementedChoreoServer) mustEmbedUnimplementedChoreoServer() {}

// UnsafeChoreoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChoreoServer will
// result in compilation errors.
type UnsafeChoreoServer interface {
	mustEmbedUnimplementedChoreoServer()
}

func RegisterChoreoServer(s grpc.ServiceRegistrar, srv ChoreoServer) {
	s.RegisterService(&Choreo_ServiceDesc, srv)
}

func _Choreo_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Get_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChoreoServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/choreopb.Choreo/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChoreoServer).Get(ctx, req.(*Get_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Choreo_Apply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Apply_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChoreoServer).Apply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/choreopb.Choreo/Apply",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChoreoServer).Apply(ctx, req.(*Apply_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Choreo_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Start_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChoreoServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/choreopb.Choreo/Start",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChoreoServer).Start(ctx, req.(*Start_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Choreo_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Stop_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChoreoServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/choreopb.Choreo/Stop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChoreoServer).Stop(ctx, req.(*Stop_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Choreo_Once_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Once_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChoreoServer).Once(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/choreopb.Choreo/Once",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChoreoServer).Once(ctx, req.(*Once_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Choreo_Load_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Load_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChoreoServer).Load(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/choreopb.Choreo/Load",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChoreoServer).Load(ctx, req.(*Load_Request))
	}
	return interceptor(ctx, in, info, handler)
}

// Choreo_ServiceDesc is the grpc.ServiceDesc for Choreo service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Choreo_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "choreopb.Choreo",
	HandlerType: (*ChoreoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _Choreo_Get_Handler,
		},
		{
			MethodName: "Apply",
			Handler:    _Choreo_Apply_Handler,
		},
		{
			MethodName: "Start",
			Handler:    _Choreo_Start_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _Choreo_Stop_Handler,
		},
		{
			MethodName: "Once",
			Handler:    _Choreo_Once_Handler,
		},
		{
			MethodName: "Load",
			Handler:    _Choreo_Load_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "choreo.proto",
}
