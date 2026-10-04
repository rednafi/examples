use std::collections::HashMap;
use std::path::PathBuf;

use shorty::config::Config;

#[track_caller]
fn fixture(name: &str) -> String {
    let path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures")
        .join(name);
    std::fs::read_to_string(&path)
        .unwrap_or_else(|e| panic!("reading {}: {e}", path.display()))
}

#[test]
fn loads_dev_env_file() {
    let vars: HashMap<_, _> = fixture("dev.env")
        .lines()
        .filter_map(|l| l.split_once('='))
        .map(|(k, v)| (k.to_string(), v.to_string()))
        .collect();

    let cfg = Config::from_lookup(|k| vars.get(k).cloned());
    assert_eq!(cfg.addr, "127.0.0.1:9000");
}

#[test]
fn cwd_is_the_package_root() {
    let cwd = std::env::current_dir().unwrap();
    assert_eq!(cwd, PathBuf::from(env!("CARGO_MANIFEST_DIR")));
}
