use assert_cmd::Command;
use predicates::str::contains;

#[test]
fn encodes_an_id() {
    Command::new(env!("CARGO_BIN_EXE_shorty"))
        .args(["encode", "125"])
        .assert()
        .success()
        .stdout("21\n");
}

#[test]
fn rejects_garbage() {
    Command::new(env!("CARGO_BIN_EXE_shorty"))
        .args(["decode", "no!"])
        .assert()
        .failure()
        .stderr(contains("invalid code"));
}
