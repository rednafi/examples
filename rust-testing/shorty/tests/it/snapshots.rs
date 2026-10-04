use expect_test::expect;
use shorty::config::Config;
use shorty::code::encode;

#[test]
fn default_config() {
    let cfg = Config::from_lookup(|_| None);
    insta::assert_debug_snapshot!(cfg);
}

#[test]
fn encode_table() {
    let table: String = [0, 61, 62, 125, 1_000_000]
        .iter()
        .map(|&id| format!("{id:>9} -> {}\n", encode(id)))
        .collect();
    insta::assert_snapshot!(table);
}

#[test]
fn inline_with_expect_test() {
    let cfg = Config::from_lookup(|_| None);
    expect![[r#"
        Config {
            addr: "0.0.0.0:8080",
            ttl: 86400s,
        }
    "#]]
    .assert_debug_eq(&cfg);
}
