use super::*;

mod encode {
    use super::*;

    #[test]
    fn zero_is_a_single_digit() {
        assert_eq!(encode(0), "0");
    }

    #[test]
    fn uses_base62() {
        assert_eq!(encode(61), "Z");
        assert_eq!(encode(62), "10");
    }
}

mod decode {
    use super::*;

    #[test]
    fn rejects_empty_input() {
        assert_eq!(decode(""), None);
    }

    #[test]
    fn rejects_overflow() {
        assert_eq!(decode("zzzzzzzzzzzzzz"), None);
    }
}

#[test]
fn digit_value_is_private_but_testable() {
    assert_eq!(digit_value(b'Z'), Some(61));
}
