// Package writefs extends the read-only io/fs.FS with a write operation using
// the same optional-interface pattern the standard library uses for
// ReadFileFS, StatFS, and friends.
//
// Companion to https://rednafi.com/go/what-io-fs-solves/.
package writefs

import (
	"errors"
	"fmt"
	"io/fs"
	"testing/fstest"
)

// WriteFileFS is an fs.FS that can also write a file. It mirrors the shape of
// fs.ReadFileFS: embed fs.FS, add one method.
type WriteFileFS interface {
	fs.FS
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

// WriteFile writes data to name in fsys. If fsys implements WriteFileFS it uses
// that; otherwise it reports that the file system is read-only, the same way
// fs.ReadFile falls back to Open.
func WriteFile(fsys fs.FS, name string, data []byte, perm fs.FileMode) error {
	if w, ok := fsys.(WriteFileFS); ok { // (1) fast path: this fs can write
		return w.WriteFile(name, data, perm)
	}
	// (2) read-only file system: nothing sensible to do, so say so clearly
	return fmt.Errorf("write %s: %w", name, errors.ErrUnsupported)
}

// MemFS is a tiny writable, in-memory file system. Reads are served by the
// embedded fstest.MapFS; writes mutate the same map.
type MemFS struct {
	fstest.MapFS
}

// NewMemFS returns an empty writable file system.
func NewMemFS() *MemFS {
	return &MemFS{MapFS: fstest.MapFS{}}
}

// WriteFile satisfies WriteFileFS.
func (m *MemFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if !fs.ValidPath(name) {
		return &fs.PathError{Op: "write", Path: name, Err: fs.ErrInvalid}
	}
	m.MapFS[name] = &fstest.MapFile{
		Data: append([]byte(nil), data...), // (1) copy; never alias the caller's slice
		Mode: perm,
	}
	return nil
}
