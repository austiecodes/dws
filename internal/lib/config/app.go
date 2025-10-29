package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

const defaultConfigPath = "configs/app.toml"

type ServerConfig struct {
	Port        int    `toml:"port"`
	SessionName string `toml:"session_name"`
	SessionKey  string `toml:"session_key"`
	AESKey      string `toml:"aes_key"`
}

type DatabaseConfig struct {
	Host             string        `toml:"host"`
	Port             int           `toml:"port"`
	User             string        `toml:"user"`
	Password         string        `toml:"password"`
	Name             string        `toml:"name"`
	SSLMode          string        `toml:"ssl_mode"`
	MaxIdleConns     int           `toml:"max_idle_conns"`
	MaxOpenConns     int           `toml:"max_open_conns"`
	ConnMaxLifetime  time.Duration `toml:"conn_max_lifetime"`
	ConnMaxIdleTime  time.Duration `toml:"conn_max_idle_time"`
	DisableTimestamp bool          `toml:"disable_timestamp_triggers"`
}

type DockerConfig struct {
	Host              string   `toml:"host"`
	APIVersion        string   `toml:"api_version"`
	SSHPortRangeStart int      `toml:"ssh_port_range_start"`
	SSHPortRangeEnd   int      `toml:"ssh_port_range_end"`
	AllowedImages     []string `toml:"allowed_images"`
	Network           string   `toml:"network"`
}

type AppConfig struct {
	App      ServerConfig   `toml:"app"`
	Database DatabaseConfig `toml:"database"`
	Docker   DockerConfig   `toml:"docker"`
}

// Load reads the TOML configuration from the provided path. If the path is empty,
// it falls back to configs/app.toml unless DWS_CONFIG_PATH is set.
func Load(path string) (*AppConfig, error) {
	configPath := path
	if configPath == "" {
		if fromEnv := os.Getenv("DWS_CONFIG_PATH"); fromEnv != "" {
			configPath = fromEnv
		} else {
			configPath = defaultConfigPath
		}
	}

	var cfg AppConfig
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("load config (%s): %w", configPath, err)
	}
	cfg.Database.applyDefaults()
	cfg.Docker.applyDefaults()
	return &cfg, nil
}

func (d *DatabaseConfig) applyDefaults() {
	if d.SSLMode == "" {
		d.SSLMode = "disable"
	}
	if d.MaxIdleConns == 0 {
		d.MaxIdleConns = 10
	}
	if d.MaxOpenConns == 0 {
		d.MaxOpenConns = 25
	}
	if d.ConnMaxLifetime == 0 {
		d.ConnMaxLifetime = 30 * time.Minute
	}
	if d.ConnMaxIdleTime == 0 {
		d.ConnMaxIdleTime = 10 * time.Minute
	}
}

func (d *DockerConfig) applyDefaults() {
	if d.SSHPortRangeStart == 0 {
		d.SSHPortRangeStart = 22000
	}
	if d.SSHPortRangeEnd == 0 || d.SSHPortRangeEnd < d.SSHPortRangeStart {
		d.SSHPortRangeEnd = d.SSHPortRangeStart + 999
	}
}
