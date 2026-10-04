use proptest::prelude::*;
use shorty::code::{decode, encode};

proptest! {
    #[test]
    fn decode_inverts_encode(id: u64) {
        prop_assert_eq!(decode(&encode(id)), Some(id));
    }

    #[test]
    #[ignore = "deliberately false, shown in the talk"]
    fn codes_fit_in_ten_chars(id: u64) {
        prop_assert!(encode(id).len() <= 10);
    }
}
