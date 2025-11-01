package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

const (
	defaultWorkerConfigPath    = "configs/worker.toml"
	workerConfigEnvOverrideKey = "DWS_WORKER_CONFIG_PATH"
)

type WorkerRPCConfig struct {
	ListenAddr    string `toml:"listen_addr"`
	SchedulerAddr string `toml:"scheduler_addr"`
}

type WorkerNodeConfig struct {
	ID               string          `toml:"id"`
	Name             string          `toml:"name"`
	Address          string          `toml:"address"`
	MaxTasks         int             `toml:"max_tasks"`
	HeartbeatTTLSecs int             `toml:"heartbeat_ttl_secs"`
	RPC              WorkerRPCConfig `toml:"rpc"`
}

type WorkerConfig struct {
	Database DatabaseConfig   `toml:"database"`
	Docker   DockerConfig     `toml:"docker"`
	Worker   WorkerNodeConfig `toml:"worker"`
}

func LoadWorker(path string) (*WorkerConfig, error) {
	configPath := resolveConfigPath(path, workerConfigEnvOverrideKey, defaultWorkerConfigPath)

	var cfg WorkerConfig
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("load worker config (%s): %w", configPath, err)
	}
	cfg.Database.applyDefaults()
	cfg.Docker.applyDefaults()
	cfg.Worker.applyDefaults()

	if cfg.Worker.ID == "" {
		return nil, fmt.Errorf("worker id must be specified in %s", configPath)
	}

	return &cfg, nil
}

func (c *WorkerRPCConfig) applyDefaults() {
	if c.ListenAddr == "" {
		c.ListenAddr = "0.0.0.0:7002"
	}
	if c.SchedulerAddr == "" {
		c.SchedulerAddr = "127.0.0.1:7001"
	}
}

func (w *WorkerNodeConfig) applyDefaults() {
	w.RPC.applyDefaults()
	if w.MaxTasks <= 0 {
		w.MaxTasks = 2
	}
	if w.Address == "" {
		w.Address = w.RPC.ListenAddr
	}
	if w.Name == "" {
		w.Name = w.ID
	}
	if w.HeartbeatTTLSecs <= 0 {
		w.HeartbeatTTLSecs = 30
	}
}
