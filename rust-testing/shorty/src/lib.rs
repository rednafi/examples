//! A tiny URL shortener used as the running example.

pub mod code;
pub mod config;
pub mod export;
pub mod health;
pub mod http;
pub mod idgen;
pub mod service;
pub mod service_dyn;
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
        let result = add(2, 2);
        assert_eq!(result, 4);
    }
}
