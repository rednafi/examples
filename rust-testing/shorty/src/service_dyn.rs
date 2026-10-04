//! The same service as `service.rs`, holding a trait object instead of a type parameter.

use std::sync::Arc;

use crate::code::encode;
use crate::idgen::IdGen;
use crate::store::Store;

pub struct DynShortener {
    store: Arc<dyn Store>,
    ids: IdGen,
}

impl DynShortener {
    pub fn new(store: Arc<dyn Store>) -> Self {
        Self { store, ids: IdGen::starting_at(1000) }
    }

    pub fn shorten(&self, url: &str) -> String {
        let code = encode(self.ids.next());
        self.store.put(&code, url);
        code
    }

    pub fn resolve(&self, code: &str) -> Option<String> {
        self.store.get(code)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::store::{MemStore, MockStore};

    #[test]
    fn works_with_any_store_chosen_at_runtime() {
        let svc = DynShortener::new(Arc::new(MemStore::default()));
        let code = svc.shorten("https://rust-lang.org");
        assert_eq!(svc.resolve(&code).as_deref(), Some("https://rust-lang.org"));
    }

    #[test]
    fn a_mock_fits_behind_dyn_too() {
        let mut store = MockStore::new();
        store.expect_get().returning(|_| None);

        let svc = DynShortener::new(Arc::new(store));
        assert_eq!(svc.resolve("g8"), None);
    }
}
