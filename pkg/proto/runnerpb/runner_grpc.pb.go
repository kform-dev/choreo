// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.24.3
// source: runner.proto

package runnerpb

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

// RunnerClient is the client API for Runner service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RunnerClient interface {
	Start(ctx context.Context, in *Start_Request, opts ...grpc.CallOption) (*Start_Response, error)
	Stop(ctx context.Context, in *Stop_Request, opts ...grpc.CallOption) (*Stop_Response, error)
	Once(ctx context.Context, in *Once_Request, opts ...grpc.CallOption) (*Once_Response, error)
	Load(ctx context.Context, in *Load_Request, opts ...grpc.CallOption) (*Load_Response, error)
}

type runnerClient struct {
	cc grpc.ClientConnInterface
}

func NewRunnerClient(cc grpc.ClientConnInterface) RunnerClient {
	return &runnerClient{cc}
}

func (c *runnerClient) Start(ctx context.Context, in *Start_Request, opts ...grpc.CallOption) (*Start_Response, error) {
	out := new(Start_Response)
	err := c.cc.Invoke(ctx, "/runnerpb.Runner/Start", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runnerClient) Stop(ctx context.Context, in *Stop_Request, opts ...grpc.CallOption) (*Stop_Response, error) {
	out := new(Stop_Response)
	err := c.cc.Invoke(ctx, "/runnerpb.Runner/Stop", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runnerClient) Once(ctx context.Context, in *Once_Request, opts ...grpc.CallOption) (*Once_Response, error) {
	out := new(Once_Response)
	err := c.cc.Invoke(ctx, "/runnerpb.Runner/Once", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *runnerClient) Load(ctx context.Context, in *Load_Request, opts ...grpc.CallOption) (*Load_Response, error) {
	out := new(Load_Response)
	err := c.cc.Invoke(ctx, "/runnerpb.Runner/Load", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RunnerServer is the server API for Runner service.
// All implementations must embed UnimplementedRunnerServer
// for forward compatibility
type RunnerServer interface {
	Start(context.Context, *Start_Request) (*Start_Response, error)
	Stop(context.Context, *Stop_Request) (*Stop_Response, error)
	Once(context.Context, *Once_Request) (*Once_Response, error)
	Load(context.Context, *Load_Request) (*Load_Response, error)
	mustEmbedUnimplementedRunnerServer()
}

// UnimplementedRunnerServer must be embedded to have forward compatible implementations.
type UnimplementedRunnerServer struct {
}

func (UnimplementedRunnerServer) Start(context.Context, *Start_Request) (*Start_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Start not implemented")
}
func (UnimplementedRunnerServer) Stop(context.Context, *Stop_Request) (*Stop_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
func (UnimplementedRunnerServer) Once(context.Context, *Once_Request) (*Once_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Once not implemented")
}
func (UnimplementedRunnerServer) Load(context.Context, *Load_Request) (*Load_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Load not implemented")
}
func (UnimplementedRunnerServer) mustEmbedUnimplementedRunnerServer() {}

// UnsafeRunnerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RunnerServer will
// result in compilation errors.
type UnsafeRunnerServer interface {
	mustEmbedUnimplementedRunnerServer()
}

func RegisterRunnerServer(s grpc.ServiceRegistrar, srv RunnerServer) {
	s.RegisterService(&Runner_ServiceDesc, srv)
}

func _Runner_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Start_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/runnerpb.Runner/Start",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).Start(ctx, req.(*Start_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Runner_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Stop_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/runnerpb.Runner/Stop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).Stop(ctx, req.(*Stop_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Runner_Once_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Once_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).Once(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/runnerpb.Runner/Once",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).Once(ctx, req.(*Once_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Runner_Load_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Load_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerServer).Load(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/runnerpb.Runner/Load",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerServer).Load(ctx, req.(*Load_Request))
	}
	return interceptor(ctx, in, info, handler)
}

// Runner_ServiceDesc is the grpc.ServiceDesc for Runner service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Runner_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "runnerpb.Runner",
	HandlerType: (*RunnerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Start",
			Handler:    _Runner_Start_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _Runner_Stop_Handler,
		},
		{
			MethodName: "Once",
			Handler:    _Runner_Once_Handler,
		},
		{
			MethodName: "Load",
			Handler:    _Runner_Load_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "runner.proto",
}