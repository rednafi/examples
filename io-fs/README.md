## io-fs

Companion code for
[What does Go's io/fs actually solve?](https://rednafi.com/go/what-io-fs-solves/).

One config loader, `config.Load(fsys fs.FS, name string)`, runs against every
file system in here without changing a line.

- `config/` — the loader plus two tests of the same thing: an in-memory
  `fstest.MapFS` and a `t.TempDir` on the real disk.
- `anyfs/` — a runnable program that loads the same config from `os.DirFS`,
  `embed.FS`, an in-memory `zip.Reader`, and `fstest.MapFS`.
- `fixtures/` — a whole directory tree kept as one `txtar` text block, handed to
  the loader through `txtar.FS`, and walked with `fs.WalkDir`.
- `golden/` — golden-file testing with fixtures embedded via `embed.FS`.
- `writefs/` — a `WriteFileFS` extension interface that adds writes the same way
  the standard library adds `ReadFileFS`, plus `os.CopyFS` and `fstest.TestFS`.

### Run the tests

```sh
go test ./...
```

### Run the four-backend demo

```sh
cd anyfs && go run .
```
