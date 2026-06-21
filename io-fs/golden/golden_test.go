// Package golden shows golden-file testing with embedded fixtures. The golden
// files live in testdata/ and ride along in the test binary via embed.FS, so
// the test reads them through io/fs instead of touching the disk at run time.
package golden

import (
	"embed"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/*.golden
var golden embed.FS

// render is the function under test.
func render(name string) string {
	return "Hello, " + name + "!\n"
}

var update = flag.Bool("update", false, "rewrite golden files")

func TestRender(t *testing.T) {
	got := render("fs")

	const name = "testdata/greeting.golden"
	if *update { // (1) regenerate with: go test -update
		if err := os.WriteFile(filepath.FromSlash(name), []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	want, err := fs.ReadFile(golden, name) // (2) read the embedded golden file
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("render mismatch:\n got: %q\nwant: %q", got, want)
	}
}
