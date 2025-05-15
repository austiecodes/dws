package clients

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

var (
	projectRoot string
	rootOnce    sync.Once
)

// InitProjectRoot initializes the project root directory
func InitProjectRoot() error {
	var initErr error
	rootOnce.Do(func() {
		// First try environment variable
		projectRoot = os.Getenv("DWS_PROJECT_ROOT")
		if projectRoot != "" {
			return
		}

		// If not set, try to find the project root
		dir, err := os.Getwd()
		if err != nil {
			initErr = fmt.Errorf("failed to get working directory: %w", err)
			return
		}

		// Look for the 'conf' directory to identify project root
		for {
			if _, err := os.Stat(filepath.Join(dir, "conf")); err == nil {
				projectRoot = dir
				return
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				initErr = fmt.Errorf("could not find project root (no conf directory found)")
				return
			}
			dir = parent
		}
	})
	return initErr
}

// GetProjectRoot returns the project root directory
func GetProjectRoot() string {
	return projectRoot
}

// LoadConfig loads a TOML configuration file from the conf directory
func LoadConfig(filename string, config interface{}) error {
	if projectRoot == "" {
		if err := InitProjectRoot(); err != nil {
			return fmt.Errorf("project root not initialized: %w", err)
		}
	}

	configPath := filepath.Join(projectRoot, "conf", filename)
	if _, err := toml.DecodeFile(configPath, config); err != nil {
		return fmt.Errorf("error loading config from %s: %w", configPath, err)
	}

	return nil
}

// ResolvePath converts a relative path to absolute path based on project root
func ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(projectRoot, path)
}
