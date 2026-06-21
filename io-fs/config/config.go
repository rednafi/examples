// Package config loads a small JSON config from any io/fs file system.
//
// The loader takes an fs.FS, so the same code reads from the real disk,
// embedded files, a zip, an in-memory map, or a txtar archive.
//
// Companion to https://rednafi.com/go/what-io-fs-solves/.
package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// Config is the decoded application config.
type Config struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// Load reads name from fsys and decodes it as JSON.
func Load(fsys fs.FS, name string) (*Config, error) {
	data, err := fs.ReadFile(fsys, name) // (1) one read, works on any backend
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", name, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil { // (2) decode the bytes
		return nil, fmt.Errorf("parse config %q: %w", name, err)
	}
	return &cfg, nil
}
