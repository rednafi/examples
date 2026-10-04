use std::collections::HashMap;
use std::sync::Mutex;

#[cfg_attr(test, mockall::automock)]
pub trait Store: Send + Sync {
    fn put(&self, code: &str, url: &str);
    fn get(&self, code: &str) -> Option<String>;
}

#[derive(Default)]
pub struct MemStore {
    links: Mutex<HashMap<String, String>>,
}

impl Store for MemStore {
    fn put(&self, code: &str, url: &str) {
        self.links.lock().unwrap().insert(code.into(), url.into());
    }

    fn get(&self, code: &str) -> Option<String> {
        self.links.lock().unwrap().get(code).cloned()
    }
}
