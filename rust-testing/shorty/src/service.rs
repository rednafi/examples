use crate::code::encode;
use crate::idgen::IdGen;
use crate::store::Store;

pub struct Shortener<S> {
    store: S,
    ids: IdGen,
}

impl<S: Store> Shortener<S> {
    pub fn new(store: S) -> Self {
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
    use mockall::predicate::eq;

    #[test]
    fn round_trips_through_a_real_fake() {
        let svc = Shortener::new(MemStore::default());
        let code = svc.shorten("https://rust-lang.org");
        assert_eq!(svc.resolve(&code).as_deref(), Some("https://rust-lang.org"));
    }

    #[test]
    fn writes_once_to_the_store() {
        let mut store = MockStore::new();
        store
            .expect_put()
            .with(eq("g8"), eq("https://rust-lang.org"))
            .times(1)
            .return_const(());

        Shortener::new(store).shorten("https://rust-lang.org");
    }
}
