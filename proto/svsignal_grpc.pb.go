// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: svsignal.proto

package svsignal

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

// SVSignalClient is the client API for SVSignal service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SVSignalClient interface {
	CreateGroup(ctx context.Context, in *Group, opts ...grpc.CallOption) (*Nothing, error)
	UpdateGroup(ctx context.Context, in *Group, opts ...grpc.CallOption) (*Nothing, error)
	DeleteGroup(ctx context.Context, in *Group, opts ...grpc.CallOption) (*Nothing, error)
	GetAllGroup(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*Groups, error)
	GetGroup(ctx context.Context, in *MsgKey, opts ...grpc.CallOption) (*Group, error)
	CreateSignal(ctx context.Context, in *Signal, opts ...grpc.CallOption) (*Nothing, error)
	UpdateSignal(ctx context.Context, in *Signal, opts ...grpc.CallOption) (*Nothing, error)
	DeleteSignal(ctx context.Context, in *Signal, opts ...grpc.CallOption) (*Nothing, error)
	GetSignals(ctx context.Context, in *MsgKey, opts ...grpc.CallOption) (*Signals, error)
	GetSignal(ctx context.Context, in *MsgKey, opts ...grpc.CallOption) (*Signal, error)
	CreateTag(ctx context.Context, in *SignalTag, opts ...grpc.CallOption) (*Nothing, error)
	UpdateTag(ctx context.Context, in *SignalTag, opts ...grpc.CallOption) (*Nothing, error)
	DeleteTag(ctx context.Context, in *SignalTag, opts ...grpc.CallOption) (*Nothing, error)
	GetData(ctx context.Context, in *RequestData, opts ...grpc.CallOption) (*ResponseData, error)
	SaveValue(ctx context.Context, in *MsgSaveValue, opts ...grpc.CallOption) (*Nothing, error)
}

type sVSignalClient struct {
	cc grpc.ClientConnInterface
}

func NewSVSignalClient(cc grpc.ClientConnInterface) SVSignalClient {
	return &sVSignalClient{cc}
}

func (c *sVSignalClient) CreateGroup(ctx context.Context, in *Group, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/CreateGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) UpdateGroup(ctx context.Context, in *Group, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/UpdateGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) DeleteGroup(ctx context.Context, in *Group, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/DeleteGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) GetAllGroup(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*Groups, error) {
	out := new(Groups)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/GetAllGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) GetGroup(ctx context.Context, in *MsgKey, opts ...grpc.CallOption) (*Group, error) {
	out := new(Group)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/GetGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) CreateSignal(ctx context.Context, in *Signal, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/CreateSignal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) UpdateSignal(ctx context.Context, in *Signal, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/UpdateSignal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) DeleteSignal(ctx context.Context, in *Signal, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/DeleteSignal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) GetSignals(ctx context.Context, in *MsgKey, opts ...grpc.CallOption) (*Signals, error) {
	out := new(Signals)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/GetSignals", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) GetSignal(ctx context.Context, in *MsgKey, opts ...grpc.CallOption) (*Signal, error) {
	out := new(Signal)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/GetSignal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) CreateTag(ctx context.Context, in *SignalTag, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/CreateTag", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) UpdateTag(ctx context.Context, in *SignalTag, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/UpdateTag", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) DeleteTag(ctx context.Context, in *SignalTag, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/DeleteTag", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) GetData(ctx context.Context, in *RequestData, opts ...grpc.CallOption) (*ResponseData, error) {
	out := new(ResponseData)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/GetData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sVSignalClient) SaveValue(ctx context.Context, in *MsgSaveValue, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/svsignal.SVSignal/SaveValue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SVSignalServer is the server API for SVSignal service.
// All implementations should embed UnimplementedSVSignalServer
// for forward compatibility
type SVSignalServer interface {
	CreateGroup(context.Context, *Group) (*Nothing, error)
	UpdateGroup(context.Context, *Group) (*Nothing, error)
	DeleteGroup(context.Context, *Group) (*Nothing, error)
	GetAllGroup(context.Context, *Nothing) (*Groups, error)
	GetGroup(context.Context, *MsgKey) (*Group, error)
	CreateSignal(context.Context, *Signal) (*Nothing, error)
	UpdateSignal(context.Context, *Signal) (*Nothing, error)
	DeleteSignal(context.Context, *Signal) (*Nothing, error)
	GetSignals(context.Context, *MsgKey) (*Signals, error)
	GetSignal(context.Context, *MsgKey) (*Signal, error)
	CreateTag(context.Context, *SignalTag) (*Nothing, error)
	UpdateTag(context.Context, *SignalTag) (*Nothing, error)
	DeleteTag(context.Context, *SignalTag) (*Nothing, error)
	GetData(context.Context, *RequestData) (*ResponseData, error)
	SaveValue(context.Context, *MsgSaveValue) (*Nothing, error)
}

// UnimplementedSVSignalServer should be embedded to have forward compatible implementations.
type UnimplementedSVSignalServer struct {
}

func (UnimplementedSVSignalServer) CreateGroup(context.Context, *Group) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateGroup not implemented")
}
func (UnimplementedSVSignalServer) UpdateGroup(context.Context, *Group) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGroup not implemented")
}
func (UnimplementedSVSignalServer) DeleteGroup(context.Context, *Group) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteGroup not implemented")
}
func (UnimplementedSVSignalServer) GetAllGroup(context.Context, *Nothing) (*Groups, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllGroup not implemented")
}
func (UnimplementedSVSignalServer) GetGroup(context.Context, *MsgKey) (*Group, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGroup not implemented")
}
func (UnimplementedSVSignalServer) CreateSignal(context.Context, *Signal) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSignal not implemented")
}
func (UnimplementedSVSignalServer) UpdateSignal(context.Context, *Signal) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSignal not implemented")
}
func (UnimplementedSVSignalServer) DeleteSignal(context.Context, *Signal) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteSignal not implemented")
}
func (UnimplementedSVSignalServer) GetSignals(context.Context, *MsgKey) (*Signals, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSignals not implemented")
}
func (UnimplementedSVSignalServer) GetSignal(context.Context, *MsgKey) (*Signal, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSignal not implemented")
}
func (UnimplementedSVSignalServer) CreateTag(context.Context, *SignalTag) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTag not implemented")
}
func (UnimplementedSVSignalServer) UpdateTag(context.Context, *SignalTag) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTag not implemented")
}
func (UnimplementedSVSignalServer) DeleteTag(context.Context, *SignalTag) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTag not implemented")
}
func (UnimplementedSVSignalServer) GetData(context.Context, *RequestData) (*ResponseData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetData not implemented")
}
func (UnimplementedSVSignalServer) SaveValue(context.Context, *MsgSaveValue) (*Nothing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SaveValue not implemented")
}

// UnsafeSVSignalServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SVSignalServer will
// result in compilation errors.
type UnsafeSVSignalServer interface {
	mustEmbedUnimplementedSVSignalServer()
}

func RegisterSVSignalServer(s grpc.ServiceRegistrar, srv SVSignalServer) {
	s.RegisterService(&SVSignal_ServiceDesc, srv)
}

func _SVSignal_CreateGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Group)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).CreateGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/CreateGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).CreateGroup(ctx, req.(*Group))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_UpdateGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Group)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).UpdateGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/UpdateGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).UpdateGroup(ctx, req.(*Group))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_DeleteGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Group)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).DeleteGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/DeleteGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).DeleteGroup(ctx, req.(*Group))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_GetAllGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Nothing)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).GetAllGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/GetAllGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).GetAllGroup(ctx, req.(*Nothing))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_GetGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).GetGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/GetGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).GetGroup(ctx, req.(*MsgKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_CreateSignal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Signal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).CreateSignal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/CreateSignal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).CreateSignal(ctx, req.(*Signal))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_UpdateSignal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Signal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).UpdateSignal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/UpdateSignal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).UpdateSignal(ctx, req.(*Signal))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_DeleteSignal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Signal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).DeleteSignal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/DeleteSignal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).DeleteSignal(ctx, req.(*Signal))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_GetSignals_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).GetSignals(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/GetSignals",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).GetSignals(ctx, req.(*MsgKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_GetSignal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).GetSignal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/GetSignal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).GetSignal(ctx, req.(*MsgKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_CreateTag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignalTag)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).CreateTag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/CreateTag",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).CreateTag(ctx, req.(*SignalTag))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_UpdateTag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignalTag)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).UpdateTag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/UpdateTag",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).UpdateTag(ctx, req.(*SignalTag))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_DeleteTag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignalTag)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).DeleteTag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/DeleteTag",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).DeleteTag(ctx, req.(*SignalTag))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_GetData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).GetData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/GetData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).GetData(ctx, req.(*RequestData))
	}
	return interceptor(ctx, in, info, handler)
}

func _SVSignal_SaveValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSaveValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SVSignalServer).SaveValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/svsignal.SVSignal/SaveValue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SVSignalServer).SaveValue(ctx, req.(*MsgSaveValue))
	}
	return interceptor(ctx, in, info, handler)
}

// SVSignal_ServiceDesc is the grpc.ServiceDesc for SVSignal service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SVSignal_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "svsignal.SVSignal",
	HandlerType: (*SVSignalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateGroup",
			Handler:    _SVSignal_CreateGroup_Handler,
		},
		{
			MethodName: "UpdateGroup",
			Handler:    _SVSignal_UpdateGroup_Handler,
		},
		{
			MethodName: "DeleteGroup",
			Handler:    _SVSignal_DeleteGroup_Handler,
		},
		{
			MethodName: "GetAllGroup",
			Handler:    _SVSignal_GetAllGroup_Handler,
		},
		{
			MethodName: "GetGroup",
			Handler:    _SVSignal_GetGroup_Handler,
		},
		{
			MethodName: "CreateSignal",
			Handler:    _SVSignal_CreateSignal_Handler,
		},
		{
			MethodName: "UpdateSignal",
			Handler:    _SVSignal_UpdateSignal_Handler,
		},
		{
			MethodName: "DeleteSignal",
			Handler:    _SVSignal_DeleteSignal_Handler,
		},
		{
			MethodName: "GetSignals",
			Handler:    _SVSignal_GetSignals_Handler,
		},
		{
			MethodName: "GetSignal",
			Handler:    _SVSignal_GetSignal_Handler,
		},
		{
			MethodName: "CreateTag",
			Handler:    _SVSignal_CreateTag_Handler,
		},
		{
			MethodName: "UpdateTag",
			Handler:    _SVSignal_UpdateTag_Handler,
		},
		{
			MethodName: "DeleteTag",
			Handler:    _SVSignal_DeleteTag_Handler,
		},
		{
			MethodName: "GetData",
			Handler:    _SVSignal_GetData_Handler,
		},
		{
			MethodName: "SaveValue",
			Handler:    _SVSignal_SaveValue_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "svsignal.proto",
}