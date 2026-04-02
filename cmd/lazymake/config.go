package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFile = ".lazymake.json"

// watchConfig is the root of .lazymake.json.
type watchConfig struct {
	Targets map[string]targetConfig `json:"targets"`
}

// targetConfig holds per-target settings.
type targetConfig struct {
	Watch string `json:"watch,omitempty"`
}

// loadConfig reads .lazymake.json from dir. Returns an empty config if the
// file does not exist yet; returns an error only on parse failures.
func loadConfig(dir string) (*watchConfig, error) {
	path := filepath.Join(dir, configFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &watchConfig{Targets: make(map[string]targetConfig)}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg watchConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Targets == nil {
		cfg.Targets = make(map[string]targetConfig)
	}
	return &cfg, nil
}

// saveConfig writes cfg to .lazymake.json in dir (pretty-printed).
func saveConfig(dir string, cfg *watchConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, configFile)
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// pattern returns the saved glob pattern for the named target, or "".
func (c *watchConfig) pattern(name string) string {
	return c.Targets[name].Watch
}

// setPattern stores a glob pattern for the named target and saves immediately.
func (c *watchConfig) setPattern(dir, name, pattern string) {
	tc := c.Targets[name]
	tc.Watch = pattern
	c.Targets[name] = tc
}
