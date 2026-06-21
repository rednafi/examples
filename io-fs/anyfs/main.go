// Command anyfs runs the exact same config.Load against four different file
// systems: the real disk, files baked into the binary, an in-memory zip, and
// an in-memory map. The output is identical every time.
//
// Run it from this directory so the os.DirFS example finds ./testdata:
//
//	go run .
//
// Companion to https://rednafi.com/go/what-io-fs-solves/.
package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"testing/fstest"

	"github.com/rednafi/examples/io-fs/config"
)

//go:embed testdata/config.json
var embedded embed.FS

func main() {
	const data = `{"port": 8080, "host": "localhost"}`

	backends := []struct {
		name string
		fsys fs.FS
		path string
	}{
		{"os.DirFS", os.DirFS("testdata"), "config.json"},                                  // (1) real disk
		{"embed.FS", embedded, "testdata/config.json"},                                     // (2) baked into the binary
		{"zip.Reader", mustZip("config.json", data), "config.json"},                        // (3) a zip, in memory
		{"fstest.MapFS", fstest.MapFS{"config.json": {Data: []byte(data)}}, "config.json"}, // (4) a map
	}

	for _, b := range backends {
		cfg, err := config.Load(b.fsys, b.path) // (5) one loader, four backends
		if err != nil {
			fmt.Printf("%-13s error: %v\n", b.name, err)
			continue
		}
		fmt.Printf("%-13s %+v\n", b.name, *cfg)
	}
}

// mustZip builds a one-file zip archive in memory and returns it as an fs.FS.
func mustZip(name, body string) fs.FS {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		panic(err)
	}
	if _, err := w.Write([]byte(body)); err != nil {
		panic(err)
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		panic(err)
	}
	return zr // *zip.Reader implements fs.FS
}
