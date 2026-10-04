use std::error::Error;

use rstest::{fixture, rstest};
use shorty::code::{decode, encode, pad};
use shorty::{service::Shortener, store::MemStore};

#[rstest]
#[case::zero("0", Some(0))]
#[case::last_digit("Z", Some(61))]
#[case::carry("10", Some(62))]
#[case::empty("", None)]
#[case::bad_char("a-b", None)]
fn decodes(#[case] input: &str, #[case] expected: Option<u64>) {
    assert_eq!(decode(input), expected);
}

#[fixture]
fn svc() -> Shortener<MemStore> {
    Shortener::new(MemStore::default())
}

#[rstest]
fn fresh_service_resolves_nothing(svc: Shortener<MemStore>) {
    assert_eq!(svc.resolve("g8"), None);
}

#[test]
fn round_trip() -> Result<(), Box<dyn Error>> {
    let id = decode("4c92").ok_or("decode failed")?;
    assert_eq!(encode(id), "4c92");
    Ok(())
}

#[test]
#[should_panic(expected = "exceeds max code length")]
fn pad_rejects_huge_widths() {
    pad("21", 12);
}

#[test]
#[ignore = "talks to the real internet"]
fn rust_lang_org_is_alive() {}

#[test]
#[ignore = "fails on purpose, shown in the talk"]
fn encodes_in_base62() {
    let code = encode(3843);
    assert_eq!(code, "zz", "encode(3843)");
}
