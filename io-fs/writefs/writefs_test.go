package writefs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// A writable fs goes through the fast path.
func TestWriteFile_Writable(t *testing.T) {
	fsys := NewMemFS()

	if err := WriteFile(fsys, "greeting.txt", []byte("hi"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := fs.ReadFile(fsys, "greeting.txt") // reads via the embedded MapFS
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "hi" {
		t.Fatalf("got %q", got)
	}
}

// A read-only fs (plain MapFS has no WriteFile method) reports ErrUnsupported,
// which callers can detect generically with errors.Is.
func TestWriteFile_ReadOnly(t *testing.T) {
	readonly := fstest.MapFS{"greeting.txt": {Data: []byte("hi")}}

	err := WriteFile(readonly, "greeting.txt", []byte("bye"), 0o600)
	if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("want ErrUnsupported, got %v", err)
	}
}

// ExampleWriteFile_readOnly prints the error a read-only fs produces.
func ExampleWriteFile_readOnly() {
	readonly := fstest.MapFS{"greeting.txt": {Data: []byte("hi")}}
	err := WriteFile(readonly, "greeting.txt", []byte("bye"), 0o600)
	fmt.Println(err)
	fmt.Println("is ErrUnsupported:", errors.Is(err, errors.ErrUnsupported))
	// Output:
	// write greeting.txt: unsupported operation
	// is ErrUnsupported: true
}

// MemFS is a well-behaved fs.FS, which fstest.TestFS confirms.
func TestMemFS_Conformance(t *testing.T) {
	fsys := NewMemFS()
	if err := WriteFile(fsys, "a.txt", []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(fsys, "dir/b.txt", []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := fstest.TestFS(fsys, "a.txt", "dir/b.txt"); err != nil {
		t.Fatal(err)
	}
}

// os.CopyFS is the one writer the standard library ships: it materializes any
// fs.FS onto the real disk.
func TestCopyFS(t *testing.T) {
	src := fstest.MapFS{
		"config.json":     {Data: []byte(`{"port": 8080}`)},
		"secrets/key.txt": {Data: []byte("hunter2")},
	}

	dir := t.TempDir()
	if err := os.CopyFS(dir, src); err != nil {
		t.Fatalf("CopyFS: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "secrets", "key.txt"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "hunter2" {
		t.Fatalf("got %q", got)
	}
}
