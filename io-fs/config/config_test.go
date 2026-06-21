package config

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// TestLoad_MapFS uses an in-memory file system. No disk, no cleanup; the
// fixture is a struct literal you can read at a glance.
func TestLoad_MapFS(t *testing.T) {
	fsys := fstest.MapFS{
		"config.json": {Data: []byte(`{"port": 8080, "host": "localhost"}`)},
	}

	cfg, err := Load(fsys, "config.json")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 8080 || cfg.Host != "localhost" {
		t.Fatalf("got %+v", cfg)
	}
}

// TestLoad_TempDir does the same thing the old way: a real temp directory and
// a real file. It works, but the fixture is built procedurally and every step
// can fail.
func TestLoad_TempDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"port": 8080, "host": "localhost"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := Load(os.DirFS(dir), "config.json")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 8080 || cfg.Host != "localhost" {
		t.Fatalf("got %+v", cfg)
	}
}
