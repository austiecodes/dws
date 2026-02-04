package config

import "os"

func resolveConfigPath(path, envKey, fallback string) string {
	if path != "" {
		return path
	}
	if env := os.Getenv(envKey); env != "" {
		return env
	}
	return fallback
}
