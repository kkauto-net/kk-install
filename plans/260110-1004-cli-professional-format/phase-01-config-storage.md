---
phase: 01
title: Config Storage for Language Preference
status: done
effort: 30m
reviewed: 2026-01-10
review-score: 8/10
---

# Phase 01: Config Storage

## Context

- Parent plan: [plan.md](plan.md)
- Dependencies: None

## Overview

Create `pkg/config/config.go` to manage user preferences, starting with language setting.

## Requirements

1. Store config in `~/.kk/config.yaml`
2. Auto-create directory if not exists
3. Default to English if no config
4. Thread-safe read/write

## Architecture

```go
// pkg/config/config.go
package config

type Config struct {
    Language string `yaml:"language"` // "en" or "vi"
}

func Load() (*Config, error)
func (c *Config) Save() error
func ConfigDir() string  // Returns ~/.kk
```

## Implementation Steps

### 1. Create pkg/config directory
```bash
mkdir -p pkg/config
```

### 2. Implement config.go

```go
package config

import (
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

const (
    configDirName  = ".kk"
    configFileName = "config.yaml"
)

type Config struct {
    Language string `yaml:"language"`
}

// ConfigDir returns the config directory path
func ConfigDir() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, configDirName)
}

// ConfigPath returns the full config file path
func ConfigPath() string {
    return filepath.Join(ConfigDir(), configFileName)
}

// Load reads config from disk, returns defaults if not exists
func Load() (*Config, error) {
    cfg := &Config{Language: "en"} // default

    data, err := os.ReadFile(ConfigPath())
    if err != nil {
        if os.IsNotExist(err) {
            return cfg, nil // Return default
        }
        return nil, err
    }

    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, err
    }

    return cfg, nil
}

// Save writes config to disk
func (c *Config) Save() error {
    // Create dir if not exists
    if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
        return err
    }

    data, err := yaml.Marshal(c)
    if err != nil {
        return err
    }

    return os.WriteFile(ConfigPath(), data, 0644)
}
```

### 3. Add yaml dependency

```bash
go get gopkg.in/yaml.v3
```

### 4. Update cmd/init.go to save language preference

After language selection form, add:
```go
// Save language preference
cfg, _ := config.Load()
cfg.Language = langChoice
cfg.Save()
```

### 5. Update cmd/root.go to load language on startup

In `init()`:
```go
cfg, _ := config.Load()
ui.SetLanguage(ui.Language(cfg.Language))
```

## Todo List

- [ ] Create pkg/config directory
- [ ] Implement config.go with Load/Save
- [ ] Add yaml dependency
- [ ] Update init.go to save language
- [ ] Update root.go to load language on startup
- [ ] Add unit tests

## Success Criteria

- [ ] `~/.kk/config.yaml` created after `kk init`
- [ ] Language persists between sessions
- [ ] Graceful handling of missing/corrupt config

## Files Changed

| File | Action |
|------|--------|
| `pkg/config/config.go` | CREATE |
| `cmd/init.go` | MODIFY |
| `cmd/root.go` | MODIFY |
| `go.mod` | MODIFY (add yaml dep) |
