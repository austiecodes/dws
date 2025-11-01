package grpcserver

import (
	"context"
	"errors"
	"log"
	"net"
	"strings"
	"time"

	"github.com/austiecodes/dws/internal/lib/apis/codec"
	"github.com/austiecodes/dws/internal/lib/apis/workerpb"
	libconfig "github.com/austiecodes/dws/internal/lib/config"
	libdocker "github.com/austiecodes/dws/internal/lib/docker"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// Server exposes container lifecycle controls over gRPC.
type Server struct {
	workerpb.UnimplementedWorkerControlServiceServer

	addr       string
	docker     *libdocker.Manager
	grpcServer *grpc.Server
}

// NewServer constructs a worker control RPC server.
func NewServer(cfg libconfig.WorkerNodeConfig) *Server {
	codec.Register()
	return &Server{
		addr:   cfg.RPC.ListenAddr,
		docker: libdocker.MustInstance(),
	}
}

// Start begins listening for RPC traffic.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     10 * time.Minute,
			MaxConnectionAge:      12 * time.Hour,
			MaxConnectionAgeGrace: 5 * time.Minute,
			Time:                  2 * time.Minute,
			Timeout:               20 * time.Second,
		}),
	}

	s.grpcServer = grpc.NewServer(opts...)
	workerpb.RegisterWorkerControlServiceServer(s.grpcServer, s)

	go func() {
		log.Printf("[worker] control gRPC listening on %s", s.addr)
		if err := s.grpcServer.Serve(lis); err != nil {
			if !errors.Is(err, grpc.ErrServerStopped) {
				log.Printf("[worker] control gRPC server stopped: %v", err)
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) {
	if s.grpcServer == nil {
		return
	}
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()
	select {
	case <-ctx.Done():
		s.grpcServer.Stop()
	case <-done:
	}
}

// CreateContainer provisions a new container on the worker host.
func (s *Server) CreateContainer(ctx context.Context, spec *workerpb.ContainerSpec) (*workerpb.ContainerInfo, error) {
	if spec == nil {
		return nil, status.Error(codes.InvalidArgument, "missing container spec")
	}
	if spec.Image == "" {
		return nil, status.Error(codes.InvalidArgument, "image is required")
	}
	if spec.HostSSHPort <= 0 {
		return nil, status.Error(codes.InvalidArgument, "host ssh port is required")
	}

	opts := libdocker.CreateContainerOptions{
		Name:     spec.Name,
		Image:    spec.Image,
		HostPort: int(spec.HostSSHPort),
		Password: spec.Password,
	}

	res, err := s.docker.CreateSSHContainer(ctx, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create container: %v", err)
	}

	return &workerpb.ContainerInfo{
		ContainerID: res.ID,
		Status:      "running",
		HostSSHPort: spec.HostSSHPort,
		Message:     "container created",
	}, nil
}

// StartContainer starts an existing container.
func (s *Server) StartContainer(ctx context.Context, req *workerpb.ContainerRequest) (*workerpb.ContainerInfo, error) {
	if req == nil || req.ContainerID == "" {
		return nil, status.Error(codes.InvalidArgument, "container id required")
	}
	if err := s.docker.StartContainer(ctx, req.ContainerID); err != nil {
		return nil, status.Errorf(codes.Internal, "start container: %v", err)
	}
	return &workerpb.ContainerInfo{ContainerID: req.ContainerID, Status: "running"}, nil
}

// StopContainer stops a running container.
func (s *Server) StopContainer(ctx context.Context, req *workerpb.ContainerRequest) (*workerpb.ContainerInfo, error) {
	if req == nil || req.ContainerID == "" {
		return nil, status.Error(codes.InvalidArgument, "container id required")
	}
	if err := s.docker.StopContainer(ctx, req.ContainerID); err != nil {
		return nil, status.Errorf(codes.Internal, "stop container: %v", err)
	}
	return &workerpb.ContainerInfo{ContainerID: req.ContainerID, Status: "stopped"}, nil
}

// DeleteContainer removes a container and associated resources.
func (s *Server) DeleteContainer(ctx context.Context, req *workerpb.ContainerRequest) (*workerpb.ContainerInfo, error) {
	if req == nil || req.ContainerID == "" {
		return nil, status.Error(codes.InvalidArgument, "container id required")
	}
	if err := s.docker.RemoveContainer(ctx, req.ContainerID); err != nil {
		return nil, status.Errorf(codes.Internal, "delete container: %v", err)
	}
	return &workerpb.ContainerInfo{ContainerID: req.ContainerID, Status: "deleted"}, nil
}

// ExecCommand runs a command in a container and returns the output.
func (s *Server) ExecCommand(ctx context.Context, req *workerpb.ExecCommandRequest) (*workerpb.ExecCommandResponse, error) {
	if req == nil || req.ContainerID == "" {
		return nil, status.Error(codes.InvalidArgument, "container id required")
	}
	if len(req.Command) == 0 {
		return nil, status.Error(codes.InvalidArgument, "command required")
	}

	// Join command for logging clarity
	log.Printf("[worker] exec container=%s cmd=%s", req.ContainerID, strings.Join(req.Command, " "))

	res := s.docker.ExecCommand(ctx, req.ContainerID, req.Command)
	if res.Error != nil {
		return &workerpb.ExecCommandResponse{
			ExitCode: -1,
			Output:   res.Output,
			Error:    res.Error.Error(),
		}, nil
	}

	return &workerpb.ExecCommandResponse{
		ExitCode: int32(res.ExitCode),
		Output:   res.Output,
	}, nil
}
