use std::path::PathBuf;
use std::process::Command;
use std::sync::Arc;
use std::time::Duration;

use shorty::{service::Shortener, store::MemStore};
use tokio::task::JoinSet;
use tokio::time::timeout;

use tempfile::TempDir;

#[track_caller]
fn write_env(dir: &TempDir, body: &str) -> PathBuf {
    let path = dir.path().join("shorty.env");
    std::fs::write(&path, body).expect("writing env file");
    path
}

#[test]
fn env_file_round_trip() {
    let dir = tempfile::tempdir().unwrap();
    let path = write_env(&dir, "SHORTY_TTL_SECS=60\n");

    assert!(std::fs::read_to_string(&path).unwrap().contains("60"));
} // `dir` is dropped here and the directory is deleted

#[test]
fn tempdir_is_removed_on_drop() {
    let dir = tempfile::tempdir().unwrap();
    let path = dir.path().to_path_buf();
    drop(dir);
    assert!(!path.exists());
}

#[tokio::test(flavor = "multi_thread", worker_threads = 2)]
async fn ids_from_many_tasks() {
    let svc = Arc::new(Shortener::new(MemStore::default()));

    let mut set = JoinSet::new();
    for _ in 0..100 {
        let svc = svc.clone();
        set.spawn(async move { svc.shorten("https://a.example") });
    }

    let codes = timeout(Duration::from_secs(5), set.join_all())
        .await
        .expect("tasks hung");
    assert_eq!(codes.len(), 100);
}

fn has_git() -> bool {
    Command::new("git").arg("--version").output().is_ok()
}

#[test]
fn reads_git_version() {
    if !has_git() {
        eprintln!("git not found, skipping");
        return;
    }
    let out = Command::new("git").arg("--version").output().unwrap();
    assert!(out.status.success());
}

#[test]
fn result_error_output() -> Result<(), String> {
    Ok(())
}
