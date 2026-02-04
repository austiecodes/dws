package schedulerclient

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/austiecodes/dws/internal/lib/apis/codec"
	"github.com/austiecodes/dws/internal/lib/apis/schedulerpb"
	libconfig "github.com/austiecodes/dws/internal/lib/config"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/worker"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client handles the scheduler RPC lifecycle for a worker node.
type Client struct {
	cfg      libconfig.WorkerNodeConfig
	executor *worker.Executor

	conn       *grpc.ClientConn
	scheduler  schedulerpb.SchedulerServiceClient
	callOpts   []grpc.CallOption
	retryDelay time.Duration

	once sync.Once
}

// New constructs a scheduler RPC client.
func New(cfg libconfig.WorkerNodeConfig, executor *worker.Executor) (*Client, error) {
	codec.Register()
	conn, err := grpc.NewClient(cfg.RPC.SchedulerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(codec.JSONCodecName)))
	if err != nil {
		return nil, err
	}

	return &Client{
		cfg:        cfg,
		executor:   executor,
		conn:       conn,
		scheduler:  schedulerpb.NewSchedulerServiceClient(conn),
		callOpts:   []grpc.CallOption{grpc.CallContentSubtype(codec.JSONCodecName)},
		retryDelay: 3 * time.Second,
	}, nil
}

// Close tears down the underlying connection.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Start spawns goroutines that manage polling and heartbeats until ctx is cancelled.
func (c *Client) Start(ctx context.Context) {
	c.once.Do(func() {
		go c.runHeartbeats(ctx)
		go c.runPollLoop(ctx)
	})
}

func (c *Client) runHeartbeats(ctx context.Context) {
	retry := c.retryDelay
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		stream, err := c.scheduler.SendHeartbeat(ctx, c.callOpts...)
		if err != nil {
			log.Printf("[worker] heartbeat stream error: %v", err)
			time.Sleep(retry)
			continue
		}

		ticker := time.NewTicker(10 * time.Second)
		err = c.sendHeartbeats(ctx, stream, ticker)
		ticker.Stop()
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("[worker] heartbeat loop ended: %v", err)
		}
		time.Sleep(retry)
	}
}

// SendOfflineOnce sends a one-shot offline heartbeat to mark this worker
// immediately offline by setting a short/zero lease on the scheduler side.
func (c *Client) SendOfflineOnce(ctx context.Context) error {
	stream, err := c.scheduler.SendHeartbeat(ctx, c.callOpts...)
	if err != nil {
		return err
	}
	hb := &schedulerpb.Heartbeat{
		WorkerID:     c.cfg.ID,
		Status:       string(libdb.WorkerStatusOffline),
		RunningTasks: c.executor.ActiveTaskCount(),
		Capacity:     int32(c.cfg.MaxTasks),
		TimestampMs:  time.Now().UnixMilli(),
		TtlSecs:      int32(c.cfg.HeartbeatTTLSecs),
		Metadata: map[string]string{
			"address": c.cfg.Address,
			"name":    c.cfg.Name,
		},
	}
	if err := stream.Send(hb); err != nil {
		return err
	}
	// best-effort close
	_, _ = stream.CloseAndRecv()
	return nil
}

func (c *Client) sendHeartbeats(ctx context.Context, stream schedulerpb.SchedulerService_SendHeartbeatClient, ticker *time.Ticker) error {
	send := func() error {
		hb := &schedulerpb.Heartbeat{
			WorkerID:     c.cfg.ID,
			Status:       string(libdb.WorkerStatusOnline),
			RunningTasks: c.executor.ActiveTaskCount(),
			Capacity:     int32(c.cfg.MaxTasks),
			TimestampMs:  time.Now().UnixMilli(),
			TtlSecs:      int32(c.cfg.HeartbeatTTLSecs),
			Metadata: map[string]string{
				"address": c.cfg.Address,
				"name":    c.cfg.Name,
			},
		}
		return stream.Send(hb)
	}

	if err := send(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			_, err := stream.CloseAndRecv()
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			return ctx.Err()
		case <-ticker.C:
			if err := send(); err != nil {
				return err
			}
		}
	}
}

func (c *Client) runPollLoop(ctx context.Context) {
	retry := c.retryDelay
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		stream, err := c.scheduler.PollTasks(ctx, c.callOpts...)
		if err != nil {
			log.Printf("[worker] poll stream error: %v", err)
			time.Sleep(retry)
			continue
		}

		if err := c.handlePollStream(ctx, stream); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("[worker] poll loop exited: %v", err)
			time.Sleep(retry)
		}
	}
}

func (c *Client) handlePollStream(ctx context.Context, stream schedulerpb.SchedulerService_PollTasksClient) error {
	sendErr := make(chan error, 1)
	recvErr := make(chan error, 1)

	go func() {
		sendErr <- c.sendPollRequests(ctx, stream)
	}()

	go func() {
		recvErr <- c.receiveAssignments(ctx, stream)
	}()

	select {
	case err := <-sendErr:
		_ = stream.CloseSend()
		return err
	case err := <-recvErr:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) sendPollRequests(ctx context.Context, stream schedulerpb.SchedulerService_PollTasksClient) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	send := func() error {
		req := &schedulerpb.WorkerPollRequest{
			WorkerID:     c.cfg.ID,
			MaxTasks:     int32(c.cfg.MaxTasks),
			RunningTasks: c.executor.ActiveTaskCount(),
			Metadata: map[string]string{
				"address": c.cfg.Address,
			},
		}
		return stream.Send(req)
	}

	if err := send(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := send(); err != nil {
				return err
			}
		}
	}
}

func (c *Client) receiveAssignments(ctx context.Context, stream schedulerpb.SchedulerService_PollTasksClient) error {
	for {
		assignment, err := stream.Recv()
		if err != nil {
			return err
		}
		if assignment == nil {
			continue
		}

		go c.handleAssignment(ctx, assignment)
	}
}

func (c *Client) handleAssignment(ctx context.Context, assignment *schedulerpb.TaskAssignment) {
	result := c.executor.Execute(ctx, assignment)
	resultCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := c.scheduler.ReportTaskResult(resultCtx, result, c.callOpts...); err != nil {
		log.Printf("[worker] report result for task %d: %v", assignment.TaskID, err)
	}
}
