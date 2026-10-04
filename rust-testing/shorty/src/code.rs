const ALPHABET: &[u8] = b"0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";

/// Turns a numeric id into a short code.
///
/// ```
/// use shorty::code::encode;
///
/// assert_eq!(encode(125), "21");
/// ```
pub fn encode(mut id: u64) -> String {
    if id == 0 {
        return "0".into();
    }
    let mut out = Vec::new();
    while id > 0 {
        out.push(ALPHABET[(id % 62) as usize]);
        id /= 62;
    }
    out.reverse();
    String::from_utf8(out).unwrap()
}

/// Turns a short code back into an id.
///
/// ```
/// # use shorty::code::decode;
/// assert_eq!(decode("21"), Some(125));
/// assert_eq!(decode("not valid!"), None);
/// ```
pub fn decode(code: &str) -> Option<u64> {
    if code.is_empty() {
        return None;
    }
    code.bytes().try_fold(0u64, |acc, b| {
        let digit = digit_value(b)? as u64;
        acc.checked_mul(62)?.checked_add(digit)
    })
}

/// Left-pads a code with zeros.
pub fn pad(code: &str, width: usize) -> String {
    assert!(width <= 11, "width {width} exceeds max code length 11");
    format!("{code:0>width$}")
}

fn digit_value(b: u8) -> Option<u8> {
    ALPHABET.iter().position(|&c| c == b).map(|i| i as u8)
}

#[cfg(test)]
mod tests;
