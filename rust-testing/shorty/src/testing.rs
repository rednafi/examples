//! Test helpers for crates that depend on `shorty`.
//! Enabled with the `test-util` feature.

use crate::store::Store;

/// Behavior every `Store` implementation must have.
pub fn store_conformance(store: &impl Store) {
    assert_eq!(store.get("missing"), None);

    store.put("a", "https://a.example");
    assert_eq!(store.get("a").as_deref(), Some("https://a.example"));

    store.put("a", "https://b.example");
    assert_eq!(store.get("a").as_deref(), Some("https://b.example"));
}

#[cfg(test)]
mod tests {
    #[test]
    fn mem_store_conforms() {
        super::store_conformance(&crate::store::MemStore::default());
    }
}
