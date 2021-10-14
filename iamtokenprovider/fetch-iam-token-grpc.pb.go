// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package iamtokenprovider

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion6

// IAMTokenProviderClient is the client API for IAMTokenProvider service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IAMTokenProviderClient interface {
	// Get IAM API key
	GetIAMToken(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*IAMToken, error)
}

type iAMTokenProviderClient struct {
	cc grpc.ClientConnInterface
}

func NewIAMTokenProviderClient(cc grpc.ClientConnInterface) IAMTokenProviderClient {
	return &iAMTokenProviderClient{cc}
}

func (c *iAMTokenProviderClient) GetIAMToken(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*IAMToken, error) {
	out := new(IAMToken)
	err := c.cc.Invoke(ctx, "/iamtokenprovider.IAMTokenProvider/GetIAMToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IAMTokenProviderServer is the server API for IAMTokenProvider service.
// All implementations must embed UnimplementedIAMTokenProviderServer
// for forward compatibility
type IAMTokenProviderServer interface {
	// Get IAM API key
	GetIAMToken(context.Context, *EmptyRequest) (*IAMToken, error)
	mustEmbedUnimplementedIAMTokenProviderServer()
}

// UnimplementedIAMTokenProviderServer must be embedded to have forward compatible implementations.
type UnimplementedIAMTokenProviderServer struct {
}

func (UnimplementedIAMTokenProviderServer) GetIAMToken(context.Context, *EmptyRequest) (*IAMToken, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIAMToken not implemented")
}
func (UnimplementedIAMTokenProviderServer) mustEmbedUnimplementedIAMTokenProviderServer() {}

// UnsafeIAMTokenProviderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IAMTokenProviderServer will
// result in compilation errors.
type UnsafeIAMTokenProviderServer interface {
	mustEmbedUnimplementedIAMTokenProviderServer()
}

func RegisterIAMTokenProviderServer(s *grpc.Server, srv IAMTokenProviderServer) {
	s.RegisterService(&IAMTokenProvider_ServiceDesc, srv)
}

func _IAMTokenProvider_GetIAMToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMTokenProviderServer).GetIAMToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/iamtokenprovider.IAMTokenProvider/GetIAMToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMTokenProviderServer).GetIAMToken(ctx, req.(*EmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// IAMTokenProvider_ServiceDesc is the grpc.ServiceDesc for IAMTokenProvider service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var IAMTokenProvider_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "iamtokenprovider.IAMTokenProvider",
	HandlerType: (*IAMTokenProviderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetIAMToken",
			Handler:    _IAMTokenProvider_GetIAMToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fetch_iam_token.proto",
}