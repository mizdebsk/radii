package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultConfigPath = "/etc/radii/config.json"
)

type Config struct {
	PartnerRepos bool `json:"partner-repos"`
}

func Load(path string) (Config, error) {
	cfg := Config{}
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}
	data, err := json.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}
	return nil
}
