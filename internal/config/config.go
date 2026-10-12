// Package config resolves application paths and settings.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName = "Genetica"
	// EnvDataDir overrides the data directory (useful for tests and portable use).
	EnvDataDir = "GENETICA_DATA_DIR"
)

// Config holds runtime configuration.
type Config struct {
	DataDir  string `json:"-"`
	DBPath   string `json:"-"`
	LogDir   string `json:"-"`
	LogLevel string `json:"logLevel"`
}

// DefaultDataDir returns the per-user, platform-appropriate data directory.
// The database must never live next to the executable (permission problems).
func DefaultDataDir() (string, error) {
	if v := os.Getenv(EnvDataDir); v != "" {
		return v, nil
	}
	switch runtime.GOOS {
	case "linux":
		if x := os.Getenv("XDG_DATA_HOME"); x != "" {
			return filepath.Join(x, AppName), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", AppName), nil
	default: // windows (%AppData%), darwin (~/Library/Application Support)
		base, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(base, AppName), nil
	}
}

// Load builds the configuration, creating directories when needed.
// An empty dataDir selects the platform default.
func Load(dataDir string) (*Config, error) {
	if dataDir == "" {
		d, err := DefaultDataDir()
		if err != nil {
			return nil, fmt.Errorf("resolve data dir: %w", err)
		}
		dataDir = d
	}
	c := &Config{
		DataDir:  dataDir,
		DBPath:   filepath.Join(dataDir, "farm.db"),
		LogDir:   filepath.Join(dataDir, "logs"),
		LogLevel: "info",
	}
	for _, d := range []string{c.DataDir, c.LogDir, filepath.Join(c.DataDir, "libraries"), filepath.Join(c.DataDir, "backups")} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return nil, fmt.Errorf("create %s: %w", d, err)
		}
	}
	if err := c.loadFile(filepath.Join(dataDir, "config.json")); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) loadFile(path string) error {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(b, c); err != nil {
		return fmt.Errorf("parse config %s: %w", path, err)
	}
	return nil
}
