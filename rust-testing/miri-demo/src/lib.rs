/// Sums the first `n + 1` items without bounds checks.
pub fn sum_through(xs: &[u32], n: usize) -> u32 {
    assert!(n <= xs.len()); // bug: should be `n < xs.len()`
    let mut total = 0;
    for i in 0..=n {
        total += unsafe { *xs.as_ptr().add(i) };
    }
    total
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn sums_everything() {
        let buf = [1, 2, 0];
        assert_eq!(sum_through(&buf[..2], 2), 3);
    }
}
