package workerpb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ContainerSpec defines the desired container to create.
type ContainerSpec struct {
	Image         string            `json:"image"`
	Password      string            `json:"password"`
	Name          string            `json:"name"`
	UserID        uint64            `json:"user_id"`
	Labels        map[string]string `json:"labels,omitempty"`
	HostSSHPort   int32             `json:"host_ssh_port"`
	ContainerUUID string            `json:"container_uuid"`
}

// ContainerInfo summarises the state of a container.
type ContainerInfo struct {
	ContainerID string `json:"container_id"`
	Status      string `json:"status"`
	HostSSHPort int32  `json:"host_ssh_port"`
	Message     string `json:"message"`
}

// ContainerRequest references an existing container.
type ContainerRequest struct {
	ContainerID   string `json:"container_id"`
	ContainerUUID string `json:"container_uuid"`
}

// ExecCommandRequest holds a command to run inside a container.
type ExecCommandRequest struct {
	ContainerID string   `json:"container_id"`
	Command     []string `json:"command"`
}

// ExecCommandResponse returns command output.
type ExecCommandResponse struct {
	ExitCode int32  `json:"exit_code"`
	Output   string `json:"output"`
	Error    string `json:"error"`
}

// WorkerControlServiceClient defines the client API.
type WorkerControlServiceClient interface {
	CreateContainer(ctx context.Context, in *ContainerSpec, opts ...grpc.CallOption) (*ContainerInfo, error)
	StartContainer(ctx context.Context, in *ContainerRequest, opts ...grpc.CallOption) (*ContainerInfo, error)
	StopContainer(ctx context.Context, in *ContainerRequest, opts ...grpc.CallOption) (*ContainerInfo, error)
	DeleteContainer(ctx context.Context, in *ContainerRequest, opts ...grpc.CallOption) (*ContainerInfo, error)
	ExecCommand(ctx context.Context, in *ExecCommandRequest, opts ...grpc.CallOption) (*ExecCommandResponse, error)
}

type workerControlServiceClient struct {
	cc grpc.ClientConnInterface
}

// NewWorkerControlServiceClient constructs a new client.
func NewWorkerControlServiceClient(cc grpc.ClientConnInterface) WorkerControlServiceClient {
	return &workerControlServiceClient{cc}
}

func (c *workerControlServiceClient) CreateContainer(ctx context.Context, in *ContainerSpec, opts ...grpc.CallOption) (*ContainerInfo, error) {
	out := new(ContainerInfo)
	err := c.cc.Invoke(ctx, "/dws.api.v1.WorkerControlService/CreateContainer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerControlServiceClient) StartContainer(ctx context.Context, in *ContainerRequest, opts ...grpc.CallOption) (*ContainerInfo, error) {
	out := new(ContainerInfo)
	err := c.cc.Invoke(ctx, "/dws.api.v1.WorkerControlService/StartContainer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerControlServiceClient) StopContainer(ctx context.Context, in *ContainerRequest, opts ...grpc.CallOption) (*ContainerInfo, error) {
	out := new(ContainerInfo)
	err := c.cc.Invoke(ctx, "/dws.api.v1.WorkerControlService/StopContainer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerControlServiceClient) DeleteContainer(ctx context.Context, in *ContainerRequest, opts ...grpc.CallOption) (*ContainerInfo, error) {
	out := new(ContainerInfo)
	err := c.cc.Invoke(ctx, "/dws.api.v1.WorkerControlService/DeleteContainer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerControlServiceClient) ExecCommand(ctx context.Context, in *ExecCommandRequest, opts ...grpc.CallOption) (*ExecCommandResponse, error) {
	out := new(ExecCommandResponse)
	err := c.cc.Invoke(ctx, "/dws.api.v1.WorkerControlService/ExecCommand", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkerControlServiceServer defines the server API.
type WorkerControlServiceServer interface {
	CreateContainer(context.Context, *ContainerSpec) (*ContainerInfo, error)
	StartContainer(context.Context, *ContainerRequest) (*ContainerInfo, error)
	StopContainer(context.Context, *ContainerRequest) (*ContainerInfo, error)
	DeleteContainer(context.Context, *ContainerRequest) (*ContainerInfo, error)
	ExecCommand(context.Context, *ExecCommandRequest) (*ExecCommandResponse, error)
}

// UnimplementedWorkerControlServiceServer can be embedded to ensure forward compatibility.
type UnimplementedWorkerControlServiceServer struct{}

func (UnimplementedWorkerControlServiceServer) CreateContainer(context.Context, *ContainerSpec) (*ContainerInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateContainer not implemented")
}

func (UnimplementedWorkerControlServiceServer) StartContainer(context.Context, *ContainerRequest) (*ContainerInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartContainer not implemented")
}

func (UnimplementedWorkerControlServiceServer) StopContainer(context.Context, *ContainerRequest) (*ContainerInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopContainer not implemented")
}

func (UnimplementedWorkerControlServiceServer) DeleteContainer(context.Context, *ContainerRequest) (*ContainerInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteContainer not implemented")
}

func (UnimplementedWorkerControlServiceServer) ExecCommand(context.Context, *ExecCommandRequest) (*ExecCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecCommand not implemented")
}

// RegisterWorkerControlServiceServer registers the service implementation with a gRPC server.
func RegisterWorkerControlServiceServer(s grpc.ServiceRegistrar, srv WorkerControlServiceServer) {
	s.RegisterService(&WorkerControlService_ServiceDesc, srv)
}

// WorkerControlService_ServiceDesc exposes the gRPC service descriptor.
var WorkerControlService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dws.api.v1.WorkerControlService",
	HandlerType: (*WorkerControlServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateContainer",
			Handler:    _WorkerControlService_CreateContainer_Handler,
		},
		{
			MethodName: "StartContainer",
			Handler:    _WorkerControlService_StartContainer_Handler,
		},
		{
			MethodName: "StopContainer",
			Handler:    _WorkerControlService_StopContainer_Handler,
		},
		{
			MethodName: "DeleteContainer",
			Handler:    _WorkerControlService_DeleteContainer_Handler,
		},
		{
			MethodName: "ExecCommand",
			Handler:    _WorkerControlService_ExecCommand_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v1/worker.proto",
}

func _WorkerControlService_CreateContainer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ContainerSpec)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerControlServiceServer).CreateContainer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dws.api.v1.WorkerControlService/CreateContainer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerControlServiceServer).CreateContainer(ctx, req.(*ContainerSpec))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerControlService_StartContainer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ContainerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerControlServiceServer).StartContainer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dws.api.v1.WorkerControlService/StartContainer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerControlServiceServer).StartContainer(ctx, req.(*ContainerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerControlService_StopContainer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ContainerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerControlServiceServer).StopContainer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dws.api.v1.WorkerControlService/StopContainer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerControlServiceServer).StopContainer(ctx, req.(*ContainerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerControlService_DeleteContainer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ContainerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerControlServiceServer).DeleteContainer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dws.api.v1.WorkerControlService/DeleteContainer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerControlServiceServer).DeleteContainer(ctx, req.(*ContainerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerControlService_ExecCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecCommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerControlServiceServer).ExecCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dws.api.v1.WorkerControlService/ExecCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerControlServiceServer).ExecCommand(ctx, req.(*ExecCommandRequest))
	}
	return interceptor(ctx, in, info, handler)
}
