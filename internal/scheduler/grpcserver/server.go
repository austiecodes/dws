package grpcserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"gorm.io/datatypes"

	"github.com/austiecodes/dws/internal/lib/apis/codec"
	"github.com/austiecodes/dws/internal/lib/apis/schedulerpb"
	libconfig "github.com/austiecodes/dws/internal/lib/config"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/platform/repository"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type workerState struct {
	running   int32
	capacity  int32
	updatedAt time.Time
}

// Server hosts the SchedulerService gRPC interface.
type Server struct {
	schedulerpb.UnimplementedSchedulerServiceServer

	addr       string
	grpcServer *grpc.Server

	mu         sync.Mutex
	workers    map[string]*workerState
	defaultTTL time.Duration
}

// NewServer constructs a Scheduler RPC server bound to the provided config.
func NewServer(cfg libconfig.SchedulerConfig) *Server {
	codec.Register()
	server := &Server{
		addr:       cfg.RPC.ListenAddr,
		workers:    make(map[string]*workerState),
		defaultTTL: time.Duration(cfg.HeartbeatTTLSecs) * time.Second,
	}
	return server
}

// Start boots the gRPC server in its own goroutine.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     5 * time.Minute,
			MaxConnectionAge:      12 * time.Hour,
			MaxConnectionAgeGrace: time.Minute,
			Time:                  2 * time.Minute,
			Timeout:               20 * time.Second,
		}),
	}
	s.grpcServer = grpc.NewServer(opts...)
	schedulerpb.RegisterSchedulerServiceServer(s.grpcServer, s)

	go func() {
		log.Printf("[scheduler] gRPC listening on %s", s.addr)
		if err := s.grpcServer.Serve(lis); err != nil {
			if !errors.Is(err, grpc.ErrServerStopped) {
				log.Printf("[scheduler] gRPC server exited: %v", err)
			}
		}
	}()
	return nil
}

// Stop gracefully terminates the gRPC server.
func (s *Server) Stop(ctx context.Context) {
	if s.grpcServer == nil {
		return
	}
	ch := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(ch)
	}()
	select {
	case <-ctx.Done():
		s.grpcServer.Stop()
	case <-ch:
	}
}

// PollTasks handles worker pull loops for new assignments.
func (s *Server) PollTasks(stream schedulerpb.SchedulerService_PollTasksServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if req.WorkerID == "" {
			continue
		}

		assignments, err := s.claimTasks(req)
		if err != nil {
			log.Printf("[scheduler] claim tasks for worker %s: %v", req.WorkerID, err)
			continue
		}

		if len(assignments) == 0 {
			continue
		}

		for _, assignment := range assignments {
			if err := stream.Send(assignment); err != nil {
				return err
			}
		}
	}
}

// ReportTaskResult processes task completion status sent by workers.
func (s *Server) ReportTaskResult(ctx context.Context, result *schedulerpb.TaskResult) (*schedulerpb.Ack, error) {
	if result == nil {
		return &schedulerpb.Ack{Message: "ignored"}, nil
	}

	taskID := uint(result.TaskID)

	if result.Output != "" || result.ExitCode != 0 || result.Error != "" {
		if err := repository.TaskUpdateResult(ctx, nil, taskID, result.Output, int(result.ExitCode)); err != nil {
			log.Printf("[scheduler] update result for task %d: %v", taskID, err)
		}
	}

	status := libdb.TaskStatusCompleted
	if result.Status != "" {
		status = libdb.TaskStatus(result.Status)
	} else if result.ExitCode != 0 {
		status = libdb.TaskStatusFailed
	}

	if err := repository.TaskUpdateStatus(ctx, nil, taskID, status); err != nil {
		log.Printf("[scheduler] update status for task %d: %v", taskID, err)
	}

	s.recordWorkerProgress(result.WorkerID, -1)
	return &schedulerpb.Ack{Message: "ok"}, nil
}

// SendHeartbeat records worker heartbeat payloads.
func (s *Server) SendHeartbeat(stream schedulerpb.SchedulerService_SendHeartbeatServer) error {
	for {
		hb, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return stream.SendAndClose(&schedulerpb.Ack{Message: "ok"})
			}
			return err
		}
		s.handleHeartbeat(hb)
	}
}

func (s *Server) claimTasks(req *schedulerpb.WorkerPollRequest) ([]*schedulerpb.TaskAssignment, error) {
	ctx := context.Background()
	available := req.MaxTasks - req.RunningTasks
	if available <= 0 {
		return nil, nil
	}

	tasks, err := repository.TaskListPendingForWorker(ctx, nil, req.WorkerID, int(available))
	if err != nil {
		return nil, err
	}

	assignments := make([]*schedulerpb.TaskAssignment, 0, len(tasks))
	for _, task := range tasks {
		ok, err := repository.TaskTransitionStatus(ctx, nil, task.ID, libdb.TaskStatusPending, libdb.TaskStatusRunning)
		if err != nil {
			log.Printf("[scheduler] transition task %d: %v", task.ID, err)
			continue
		}
		if !ok {
			continue
		}

		containerID := ""
		containerUUID := ""
		containerRecordID := uint64(0)
		if task.Container != nil {
			containerID = task.Container.ContainerID
			containerUUID = task.Container.UUID
			containerRecordID = uint64(task.Container.ID)
		}

		assignments = append(assignments, &schedulerpb.TaskAssignment{
			TaskID:            uint64(task.ID),
			ContainerID:       containerID,
			Command:           task.Command,
			TaskType:          string(task.TaskType),
			ContainerRecordID: containerRecordID,
			ContainerUUID:     containerUUID,
			UserID:            uint64(task.UserID),
		})
		s.recordWorkerProgress(req.WorkerID, 1)
	}

	return assignments, nil
}

func (s *Server) handleHeartbeat(hb *schedulerpb.Heartbeat) {
	if hb == nil || hb.WorkerID == "" {
		return
	}

	metadata := map[string]interface{}{
		"running_tasks": hb.RunningTasks,
		"capacity":      hb.Capacity,
		"metadata":      hb.Metadata,
	}

	payload, err := json.Marshal(metadata)
	if err != nil {
		log.Printf("[scheduler] marshal heartbeat metadata: %v", err)
		return
	}

	// Lease-based liveness: extend lease_expires_at on every heartbeat.
	// TTL precedence: hb.TtlSecs > server default > 30s fallback.
	ttl := s.defaultTTL
	if hb.TtlSecs > 0 {
		ttl = time.Duration(hb.TtlSecs) * time.Second
	}
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	lease := time.Now().Add(ttl)
	// If worker explicitly sends offline, expire lease immediately.
	if hb.Status == string(libdb.WorkerStatusOffline) {
		lease = time.Now()
	}
	updates := map[string]interface{}{
		"metadata":           datatypes.JSON(payload),
		"last_heartbeat":     time.UnixMilli(hb.TimestampMs),
		"lease_expires_at":   lease,
		"heartbeat_ttl_secs": int(ttl / time.Second),
	}
	if hb.Status == string(libdb.WorkerStatusOffline) {
		updates["status"] = libdb.WorkerStatusOffline
	}
	if addr, ok := hb.Metadata["address"]; ok && addr != "" {
		updates["address"] = addr
	}
	if name, ok := hb.Metadata["name"]; ok && name != "" {
		updates["name"] = name
	}

	if err := repository.WorkerUpdate(context.Background(), nil, hb.WorkerID, updates); err != nil {
		log.Printf("[scheduler] update worker %s heartbeat: %v", hb.WorkerID, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.workers[hb.WorkerID]
	if !ok {
		state = &workerState{}
		s.workers[hb.WorkerID] = state
	}
	state.running = hb.RunningTasks
	state.capacity = hb.Capacity
	state.updatedAt = time.Now()
}

func (s *Server) recordWorkerProgress(workerID string, delta int32) {
	if workerID == "" || delta == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.workers[workerID]
	if !ok {
		state = &workerState{}
		s.workers[workerID] = state
	}
	state.running += delta
	state.updatedAt = time.Now()
}
