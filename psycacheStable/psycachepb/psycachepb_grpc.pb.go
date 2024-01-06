// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.3
// source: psycachepb.proto

package psycachepb

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
	PsyCache_Get_FullMethodName    = "/psycachepb.PsyCache/Get"
	PsyCache_Remove_FullMethodName = "/psycachepb.PsyCache/Remove"
)

// PsyCacheClient is the client API for PsyCache service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PsyCacheClient interface {
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Remove(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*RemoveResponse, error)
}

type psyCacheClient struct {
	cc grpc.ClientConnInterface
}

func NewPsyCacheClient(cc grpc.ClientConnInterface) PsyCacheClient {
	return &psyCacheClient{cc}
}

func (c *psyCacheClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, PsyCache_Get_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *psyCacheClient) Remove(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*RemoveResponse, error) {
	out := new(RemoveResponse)
	err := c.cc.Invoke(ctx, PsyCache_Remove_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PsyCacheServer is the server API for PsyCache service.
// All implementations must embed UnimplementedPsyCacheServer
// for forward compatibility
type PsyCacheServer interface {
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Remove(context.Context, *GetRequest) (*RemoveResponse, error)
	mustEmbedUnimplementedPsyCacheServer()
}

// UnimplementedPsyCacheServer must be embedded to have forward compatible implementations.
type UnimplementedPsyCacheServer struct {
}

func (UnimplementedPsyCacheServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedPsyCacheServer) Remove(context.Context, *GetRequest) (*RemoveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Remove not implemented")
}
func (UnimplementedPsyCacheServer) mustEmbedUnimplementedPsyCacheServer() {}

// UnsafePsyCacheServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PsyCacheServer will
// result in compilation errors.
type UnsafePsyCacheServer interface {
	mustEmbedUnimplementedPsyCacheServer()
}

func RegisterPsyCacheServer(s grpc.ServiceRegistrar, srv PsyCacheServer) {
	s.RegisterService(&PsyCache_ServiceDesc, srv)
}

func _PsyCache_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PsyCacheServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PsyCache_Get_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PsyCacheServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PsyCache_Remove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PsyCacheServer).Remove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PsyCache_Remove_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PsyCacheServer).Remove(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PsyCache_ServiceDesc is the grpc.ServiceDesc for PsyCache service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PsyCache_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "psycachepb.PsyCache",
	HandlerType: (*PsyCacheServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _PsyCache_Get_Handler,
		},
		{
			MethodName: "Remove",
			Handler:    _PsyCache_Remove_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "psycachepb.proto",
}