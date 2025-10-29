package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	libconfig "github.com/austiecodes/dws/internal/lib/config"
)

type Manager struct {
	client *client.Client
	cfg    libconfig.DockerConfig
	mu     sync.Mutex
}

type CreateContainerOptions struct {
	Name     string
	Image    string
	HostPort int
	Password string   // SSH password for root user
	Env      []string // Additional environment variables
}

type CreateContainerResult struct {
	ID string
}

var (
	globalMu sync.RWMutex
	global   *Manager
)

func Init(cfg libconfig.DockerConfig) (*Manager, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	if global != nil {
		return global, nil
	}

	manager, err := newManager(cfg)
	if err != nil {
		return nil, err
	}
	global = manager
	return global, nil
}

func Instance() (*Manager, error) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	if global == nil {
		return nil, errors.New("docker manager not initialised")
	}
	return global, nil
}

func MustInstance() *Manager {
	mgr, err := Instance()
	if err != nil {
		panic(err)
	}
	return mgr
}

func newManager(cfg libconfig.DockerConfig) (*Manager, error) {
	opts := []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}
	if cfg.Host != "" {
		opts = append(opts, client.WithHost(cfg.Host))
	}
	if cfg.APIVersion != "" {
		opts = append(opts, client.WithVersion(cfg.APIVersion))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("init docker client: %w", err)
	}

	return &Manager{client: cli, cfg: cfg}, nil
}

func (m *Manager) CreateSSHContainer(ctx context.Context, opts CreateContainerOptions) (*CreateContainerResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.ensureImage(ctx, opts.Image); err != nil {
		return nil, err
	}

	port, err := nat.NewPort("tcp", "22")
	if err != nil {
		return nil, fmt.Errorf("declare ssh port: %w", err)
	}

	bindings := nat.PortMap{
		port: []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: strconv.Itoa(opts.HostPort),
		}},
	}

	// 构建环境变量：如果指定了密码，添加 SSH_PASSWORD
	env := opts.Env
	if opts.Password != "" {
		env = append(env, fmt.Sprintf("SSH_PASSWORD=%s", opts.Password))
	}

	config := &container.Config{
		Image:        opts.Image,
		Env:          env,
		ExposedPorts: nat.PortSet{port: struct{}{}},
	}

	hostCfg := &container.HostConfig{
		PortBindings: bindings,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	resp, err := m.client.ContainerCreate(ctx, config, hostCfg, nil, nil, opts.Name)
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	if err := m.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("start container: %w", err)
	}

	return &CreateContainerResult{ID: resp.ID}, nil
}

func (m *Manager) StopContainer(ctx context.Context, id string) error {
	if err := m.client.ContainerStop(ctx, id, container.StopOptions{}); err != nil {
		if client.IsErrNotFound(err) {
			return nil // 容器已不存在，视为成功
		}
		return fmt.Errorf("stop container: %w", err)
	}
	return nil
}

func (m *Manager) StartContainer(ctx context.Context, id string) error {
	if err := m.client.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		if client.IsErrNotFound(err) {
			return fmt.Errorf("container not found")
		}
		return fmt.Errorf("start container: %w", err)
	}
	return nil
}

func (m *Manager) RemoveContainer(ctx context.Context, id string) error {
	if err := m.client.ContainerStop(ctx, id, container.StopOptions{}); err != nil {
		if !client.IsErrNotFound(err) {
			return fmt.Errorf("stop container: %w", err)
		}
	}
	if err := m.client.ContainerRemove(ctx, id, container.RemoveOptions{Force: true}); err != nil {
		if client.IsErrNotFound(err) {
			return nil // 容器已不存在，视为成功
		}
		return fmt.Errorf("remove container: %w", err)
	}
	return nil
}

type ContainerStatus struct {
	ID      string
	Status  string // "running", "stopped", "paused", "exited", "deleted"
	Running bool
	Exists  bool
}

func (m *Manager) InspectContainer(ctx context.Context, id string) (*ContainerStatus, error) {
	info, err := m.client.ContainerInspect(ctx, id)
	if err != nil {
		if client.IsErrNotFound(err) {
			return &ContainerStatus{
				ID:      id,
				Status:  "deleted",
				Running: false,
				Exists:  false,
			}, nil
		}
		return nil, fmt.Errorf("inspect container: %w", err)
	}

	status := "stopped"
	if info.State.Running {
		status = "running"
	} else if info.State.Paused {
		status = "paused"
	} else if info.State.Dead {
		status = "dead"
	} else if info.State.Restarting {
		status = "restarting"
	} else if info.State.ExitCode != 0 {
		status = "exited"
	}

	return &ContainerStatus{
		ID:      id,
		Status:  status,
		Running: info.State.Running,
		Exists:  true,
	}, nil
}

func (m *Manager) ListImages(ctx context.Context) ([]string, error) {
	images, err := m.client.ImageList(ctx, imagetypes.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	var result []string
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag != "<none>:<none>" {
				result = append(result, tag)
			}
		}
	}
	return result, nil
}

func (m *Manager) ensureImage(ctx context.Context, image string) error {
	args := filters.NewArgs(filters.Arg("reference", image))
	images, err := m.client.ImageList(ctx, imagetypes.ListOptions{Filters: args})
	if err != nil {
		return fmt.Errorf("list images: %w", err)
	}
	if len(images) > 0 {
		return nil
	}

	resp, err := m.client.ImagePull(ctx, image, imagetypes.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull image: %w", err)
	}
	defer resp.Close()
	if _, err := io.Copy(io.Discard, resp); err != nil {
		return fmt.Errorf("consume pull output: %w", err)
	}
	return nil
}

func IsPortFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
