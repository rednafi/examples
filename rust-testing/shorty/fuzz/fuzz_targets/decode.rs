#![no_main]

use libfuzzer_sys::fuzz_target;
use shorty::code::{decode, encode};

fuzz_target!(|code: &str| {
    if let Some(id) = decode(code) {
        assert_eq!(encode(id), code);
    }
});
