# rust-testing

Code for the "Testing Rust" talk. Every snippet on the slides comes from one of these
crates and runs as shown.

| crate | used for |
|---|---|
| [`shorty`](shorty) | a small URL shortener: unit, doc, integration, snapshot, compile-fail, async, HTTP, CLI, property and fuzz tests |
| [`idgen`](idgen) | an atomic id generator tested with real threads and with loom |
| [`miri-demo`](miri-demo) | an off-by-one in `unsafe` code that passes `cargo test` and fails under Miri |

## shorty

`POST /links` takes a long URL, stores it under a numeric id, and returns the id
encoded in base62. `GET /{code}` decodes the code, looks up the URL, and answers with
a `307` redirect.

```sh
cd shorty
cargo run -- serve                     # SHORTY_ADDR defaults to 0.0.0.0:8080
curl -s -d https://rust-lang.org localhost:8080/links   # g8
curl -sI localhost:8080/g8                              # 307, location: https://rust-lang.org
```

Running the tests:

```sh
cargo test                             # unit, integration, data-driven and doc tests
cargo nextest run                      # same tests, one process each
cargo test -- --ignored                # the tests that fail on purpose for the talk
cargo insta test --review              # review snapshot changes
UPDATE_EXPECT=1 cargo test             # rewrite expect-test literals
TRYBUILD=overwrite cargo test --test it ui   # rewrite compile-fail .stderr files
cargo +nightly fuzz run decode         # finds "00" vs "0" within a minute
cargo mutants -f src/code.rs           # mutation testing for the codec
```

## idgen

```sh
cd idgen
cargo test                                         # std threads
RUSTFLAGS="--cfg loom" cargo test --release        # loom models
RUSTFLAGS="--cfg loom" cargo test --release -- --ignored   # watch loom catch next_racy
cargo run --release --example race                 # the race almost never shows up with real threads
```

## miri-demo

```sh
cd miri-demo
cargo test                    # passes
cargo +nightly miri test      # reports undefined behavior
```
