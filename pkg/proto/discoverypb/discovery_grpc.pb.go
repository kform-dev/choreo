// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v5.29.0
// source: discovery.proto

package discoverypb

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

// DiscoveryClient is the client API for Discovery service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DiscoveryClient interface {
	Get(ctx context.Context, in *Get_Request, opts ...grpc.CallOption) (*Get_Response, error)
	Watch(ctx context.Context, in *Watch_Request, opts ...grpc.CallOption) (Discovery_WatchClient, error)
}

type discoveryClient struct {
	cc grpc.ClientConnInterface
}

func NewDiscoveryClient(cc grpc.ClientConnInterface) DiscoveryClient {
	return &discoveryClient{cc}
}

func (c *discoveryClient) Get(ctx context.Context, in *Get_Request, opts ...grpc.CallOption) (*Get_Response, error) {
	out := new(Get_Response)
	err := c.cc.Invoke(ctx, "/discoverypb.Discovery/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *discoveryClient) Watch(ctx context.Context, in *Watch_Request, opts ...grpc.CallOption) (Discovery_WatchClient, error) {
	stream, err := c.cc.NewStream(ctx, &Discovery_ServiceDesc.Streams[0], "/discoverypb.Discovery/Watch", opts...)
	if err != nil {
		return nil, err
	}
	x := &discoveryWatchClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Discovery_WatchClient interface {
	Recv() (*Watch_Response, error)
	grpc.ClientStream
}

type discoveryWatchClient struct {
	grpc.ClientStream
}

func (x *discoveryWatchClient) Recv() (*Watch_Response, error) {
	m := new(Watch_Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DiscoveryServer is the server API for Discovery service.
// All implementations must embed UnimplementedDiscoveryServer
// for forward compatibility
type DiscoveryServer interface {
	Get(context.Context, *Get_Request) (*Get_Response, error)
	Watch(*Watch_Request, Discovery_WatchServer) error
	mustEmbedUnimplementedDiscoveryServer()
}

// UnimplementedDiscoveryServer must be embedded to have forward compatible implementations.
type UnimplementedDiscoveryServer struct {
}

func (UnimplementedDiscoveryServer) Get(context.Context, *Get_Request) (*Get_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedDiscoveryServer) Watch(*Watch_Request, Discovery_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}
func (UnimplementedDiscoveryServer) mustEmbedUnimplementedDiscoveryServer() {}

// UnsafeDiscoveryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DiscoveryServer will
// result in compilation errors.
type UnsafeDiscoveryServer interface {
	mustEmbedUnimplementedDiscoveryServer()
}

func RegisterDiscoveryServer(s grpc.ServiceRegistrar, srv DiscoveryServer) {
	s.RegisterService(&Discovery_ServiceDesc, srv)
}

func _Discovery_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Get_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiscoveryServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/discoverypb.Discovery/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiscoveryServer).Get(ctx, req.(*Get_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Discovery_Watch_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Watch_Request)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DiscoveryServer).Watch(m, &discoveryWatchServer{stream})
}

type Discovery_WatchServer interface {
	Send(*Watch_Response) error
	grpc.ServerStream
}

type discoveryWatchServer struct {
	grpc.ServerStream
}

func (x *discoveryWatchServer) Send(m *Watch_Response) error {
	return x.ServerStream.SendMsg(m)
}

// Discovery_ServiceDesc is the grpc.ServiceDesc for Discovery service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Discovery_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "discoverypb.Discovery",
	HandlerType: (*DiscoveryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _Discovery_Get_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Watch",
			Handler:       _Discovery_Watch_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "discovery.proto",
}
