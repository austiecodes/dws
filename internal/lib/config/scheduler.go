package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

const (
	defaultSchedulerConfigPath    = "configs/scheduler.toml"
	schedulerConfigEnvOverrideKey = "DWS_SCHEDULER_CONFIG_PATH"
)

type SchedulerRPCConfig struct {
	ListenAddr string `toml:"listen_addr"`
}

type SchedulerConfig struct {
	Database         DatabaseConfig     `toml:"database"`
	RPC              SchedulerRPCConfig `toml:"rpc"`
	HeartbeatTTLSecs int                `toml:"heartbeat_ttl_secs"`
}

func LoadScheduler(path string) (*SchedulerConfig, error) {
	configPath := resolveConfigPath(path, schedulerConfigEnvOverrideKey, defaultSchedulerConfigPath)

	var cfg SchedulerConfig
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("load scheduler config (%s): %w", configPath, err)
	}
	cfg.Database.applyDefaults()
	cfg.RPC.applyDefaults()
	if cfg.HeartbeatTTLSecs <= 0 {
		cfg.HeartbeatTTLSecs = 30
	}
	return &cfg, nil
}

func (r *SchedulerRPCConfig) applyDefaults() {
	if r.ListenAddr == "" {
		r.ListenAddr = "0.0.0.0:7001"
	}
}
