// Package fixtures shows how to keep a whole directory tree as a single txtar
// text block inside a test, then hand it to code that wants an fs.FS.
package fixtures

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/rednafi/examples/io-fs/config"
	"golang.org/x/tools/txtar"
)

// One readable text block describes a two-file tree. The "-- name --" markers
// separate files; everything before the first marker is a comment.
const archive = `Two-file fixture for the config loader.
-- config.json --
{"port": 8080, "host": "localhost"}
-- secrets/token.txt --
hunter2
`

func TestTxtarFS(t *testing.T) {
	fsys, err := txtar.FS(txtar.Parse([]byte(archive))) // (1) text block -> read-only fs.FS
	if err != nil {
		t.Fatalf("txtar.FS: %v", err)
	}

	cfg, err := config.Load(fsys, "config.json") // (2) same loader, txtar backend
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("got %+v", cfg)
	}

	token, err := fs.ReadFile(fsys, "secrets/token.txt") // (3) nested paths work too
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(token) != "hunter2\n" {
		t.Fatalf("got %q", token)
	}
}

// ExampleFS_walk lists the whole txtar tree with fs.WalkDir.
func ExampleFS_walk() {
	fsys, _ := txtar.FS(txtar.Parse([]byte(archive)))

	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Printf("%-20s dir=%t\n", path, d.IsDir())
		return nil
	})
	// Output:
	// .                    dir=true
	// config.json          dir=false
	// secrets              dir=true
	// secrets/token.txt    dir=false
}
