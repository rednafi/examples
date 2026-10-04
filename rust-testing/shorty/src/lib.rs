//! A tiny URL shortener used as the running example.

pub mod code;
pub mod config;
pub mod export;
pub mod health;
pub mod http;
pub mod idgen;
pub mod service;
pub mod store;

#[cfg(any(test, feature = "test-util"))]
pub mod testing;

pub fn add(left: u64, right: u64) -> u64 {
    left + right
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_works() {
        assert_eq!(add(2, 2), 4);
    }
}
