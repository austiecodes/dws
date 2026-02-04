package schedulerpb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WorkerPollRequest informs the scheduler about a worker's capacity.
type WorkerPollRequest struct {
	WorkerID     string            `json:"worker_id"`
	MaxTasks     int32             `json:"max_tasks"`
	RunningTasks int32             `json:"running_tasks"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// TaskAssignment describes an executable task payload.
type TaskAssignment struct {
	TaskID            uint64 `json:"task_id"`
	ContainerID       string `json:"container_id"`
	Command           string `json:"command"`
	TaskType          string `json:"task_type"`
	ContainerRecordID uint64 `json:"container_record_id"`
	ContainerUUID     string `json:"container_uuid"`
	UserID            uint64 `json:"user_id"`
}

// TaskResult reports execution output back to the scheduler.
type TaskResult struct {
	TaskID   uint64 `json:"task_id"`
	WorkerID string `json:"worker_id"`
	ExitCode int32  `json:"exit_code"`
	Output   string `json:"output"`
	Error    string `json:"error"`
	Status   string `json:"status"`
}

// Heartbeat contains liveness information from a worker.
type Heartbeat struct {
	WorkerID     string            `json:"worker_id"`
	Status       string            `json:"status"`
	RunningTasks int32             `json:"running_tasks"`
	Capacity     int32             `json:"capacity"`
	TimestampMs  int64             `json:"timestamp_ms"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	TtlSecs      int32             `json:"ttl_secs,omitempty"`
}

// Ack is a generic acknowledgement.
type Ack struct {
	Message string `json:"message"`
}

// SchedulerServiceClient defines the client API for scheduler interactions.
type SchedulerServiceClient interface {
	PollTasks(ctx context.Context, opts ...grpc.CallOption) (SchedulerService_PollTasksClient, error)
	ReportTaskResult(ctx context.Context, in *TaskResult, opts ...grpc.CallOption) (*Ack, error)
	SendHeartbeat(ctx context.Context, opts ...grpc.CallOption) (SchedulerService_SendHeartbeatClient, error)
}

type schedulerServiceClient struct {
	cc grpc.ClientConnInterface
}

// NewSchedulerServiceClient constructs a SchedulerServiceClient.
func NewSchedulerServiceClient(cc grpc.ClientConnInterface) SchedulerServiceClient {
	return &schedulerServiceClient{cc}
}

// PollTasks opens a bidirectional stream used to push task assignments.
func (c *schedulerServiceClient) PollTasks(ctx context.Context, opts ...grpc.CallOption) (SchedulerService_PollTasksClient, error) {
	stream, err := c.cc.NewStream(ctx, &SchedulerService_ServiceDesc.Streams[0], "/dws.api.v1.SchedulerService/PollTasks", opts...)
	if err != nil {
		return nil, err
	}
	return &schedulerServicePollTasksClient{stream}, nil
}

func (c *schedulerServiceClient) ReportTaskResult(ctx context.Context, in *TaskResult, opts ...grpc.CallOption) (*Ack, error) {
	out := new(Ack)
	err := c.cc.Invoke(ctx, "/dws.api.v1.SchedulerService/ReportTaskResult", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SendHeartbeat opens a client-stream used for periodic heartbeats.
func (c *schedulerServiceClient) SendHeartbeat(ctx context.Context, opts ...grpc.CallOption) (SchedulerService_SendHeartbeatClient, error) {
	stream, err := c.cc.NewStream(ctx, &SchedulerService_ServiceDesc.Streams[1], "/dws.api.v1.SchedulerService/SendHeartbeat", opts...)
	if err != nil {
		return nil, err
	}
	return &schedulerServiceSendHeartbeatClient{stream}, nil
}

// SchedulerServiceServer defines the server API.
type SchedulerServiceServer interface {
	PollTasks(SchedulerService_PollTasksServer) error
	ReportTaskResult(context.Context, *TaskResult) (*Ack, error)
	SendHeartbeat(SchedulerService_SendHeartbeatServer) error
}

// UnimplementedSchedulerServiceServer can be embedded to forward compatible implementations.
type UnimplementedSchedulerServiceServer struct{}

func (UnimplementedSchedulerServiceServer) PollTasks(SchedulerService_PollTasksServer) error {
	return status.Errorf(codes.Unimplemented, "method PollTasks not implemented")
}

func (UnimplementedSchedulerServiceServer) ReportTaskResult(context.Context, *TaskResult) (*Ack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportTaskResult not implemented")
}

func (UnimplementedSchedulerServiceServer) SendHeartbeat(SchedulerService_SendHeartbeatServer) error {
	return status.Errorf(codes.Unimplemented, "method SendHeartbeat not implemented")
}

// RegisterSchedulerServiceServer registers the server with a gRPC server.
func RegisterSchedulerServiceServer(s grpc.ServiceRegistrar, srv SchedulerServiceServer) {
	s.RegisterService(&SchedulerService_ServiceDesc, srv)
}

// SchedulerService_ServiceDesc exposes the service definition.
var SchedulerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dws.api.v1.SchedulerService",
	HandlerType: (*SchedulerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReportTaskResult",
			Handler:    _SchedulerService_ReportTaskResult_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PollTasks",
			Handler:       _SchedulerService_PollTasks_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "SendHeartbeat",
			Handler:       _SchedulerService_SendHeartbeat_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "api/v1/scheduler.proto",
}

type SchedulerService_PollTasksClient interface {
	Send(*WorkerPollRequest) error
	Recv() (*TaskAssignment, error)
	grpc.ClientStream
}

type schedulerServicePollTasksClient struct {
	grpc.ClientStream
}

func (x *schedulerServicePollTasksClient) Send(m *WorkerPollRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *schedulerServicePollTasksClient) Recv() (*TaskAssignment, error) {
	m := new(TaskAssignment)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

type SchedulerService_SendHeartbeatClient interface {
	Send(*Heartbeat) error
	CloseAndRecv() (*Ack, error)
	grpc.ClientStream
}

type schedulerServiceSendHeartbeatClient struct {
	grpc.ClientStream
}

func (x *schedulerServiceSendHeartbeatClient) Send(m *Heartbeat) error {
	return x.ClientStream.SendMsg(m)
}

func (x *schedulerServiceSendHeartbeatClient) CloseAndRecv() (*Ack, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	out := new(Ack)
	if err := x.ClientStream.RecvMsg(out); err != nil {
		return nil, err
	}
	return out, nil
}

type SchedulerService_PollTasksServer interface {
	Send(*TaskAssignment) error
	Recv() (*WorkerPollRequest, error)
	grpc.ServerStream
}

type SchedulerService_SendHeartbeatServer interface {
	SendAndClose(*Ack) error
	Recv() (*Heartbeat, error)
	grpc.ServerStream
}

func _SchedulerService_PollTasks_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SchedulerServiceServer).PollTasks(&schedulerServicePollTasksServer{stream})
}

type schedulerServicePollTasksServer struct {
	grpc.ServerStream
}

func (x *schedulerServicePollTasksServer) Send(m *TaskAssignment) error {
	return x.ServerStream.SendMsg(m)
}

func (x *schedulerServicePollTasksServer) Recv() (*WorkerPollRequest, error) {
	m := new(WorkerPollRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _SchedulerService_ReportTaskResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskResult)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServiceServer).ReportTaskResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dws.api.v1.SchedulerService/ReportTaskResult",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServiceServer).ReportTaskResult(ctx, req.(*TaskResult))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchedulerService_SendHeartbeat_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SchedulerServiceServer).SendHeartbeat(&schedulerServiceSendHeartbeatServer{stream})
}

type schedulerServiceSendHeartbeatServer struct {
	grpc.ServerStream
}

func (x *schedulerServiceSendHeartbeatServer) SendAndClose(m *Ack) error {
	return x.ServerStream.SendMsg(m)
}

func (x *schedulerServiceSendHeartbeatServer) Recv() (*Heartbeat, error) {
	m := new(Heartbeat)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}
