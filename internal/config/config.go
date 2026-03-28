package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultMaxHistory    = 500
	defaultPreviewLength = 60
)

// Config represents the user-configurable goclip application settings.
type Config struct {
	MaxHistory    int    `yaml:"max_history"`
	DBPath        string `yaml:"db_path"`
	PreviewLength int    `yaml:"preview_length"`
}

type fileConfig struct {
	MaxHistory    *int    `yaml:"max_history"`
	DBPath        *string `yaml:"db_path"`
	PreviewLength *int    `yaml:"preview_length"`
}

// DefaultConfig returns the default application configuration.
func DefaultConfig() Config {
	return Config{
		MaxHistory:    defaultMaxHistory,
		DBPath:        defaultDBPath(),
		PreviewLength: defaultPreviewLength,
	}
}

// Load reads goclip configuration from disk and falls back to defaults when the file is absent.
func Load(path string) (Config, error) {
	configPath, err := resolveConfigPath(path)
	if err != nil {
		return Config{}, err
	}

	cfg := DefaultConfig()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}

		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var raw fileConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if raw.MaxHistory != nil {
		if *raw.MaxHistory < 0 {
			return Config{}, errors.New("max_history cannot be negative")
		}

		cfg.MaxHistory = *raw.MaxHistory
	}

	if raw.PreviewLength != nil {
		if *raw.PreviewLength <= 0 {
			return Config{}, errors.New("preview_length must be greater than 0")
		}

		cfg.PreviewLength = *raw.PreviewLength
	}

	if raw.DBPath != nil {
		dbPath := strings.TrimSpace(*raw.DBPath)
		if dbPath != "" {
			cfg.DBPath, err = expandPath(dbPath)
			if err != nil {
				return Config{}, err
			}
		}
	}

	return cfg, nil
}

func resolveConfigPath(path string) (string, error) {
	if strings.TrimSpace(path) != "" {
		return expandPath(path)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, ".goclip", "config.yaml"), nil
}

func defaultDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".goclip", "history.db")
	}

	return filepath.Join(homeDir, ".goclip", "history.db")
}

func expandPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}

		if path == "~" {
			return homeDir, nil
		}

		return filepath.Join(homeDir, path[2:]), nil
	}

	return path, nil
}
