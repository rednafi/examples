use std::sync::atomic::{AtomicU64, Ordering};

pub struct IdGen {
    next: AtomicU64,
}

impl IdGen {
    pub fn starting_at(n: u64) -> Self {
        Self { next: AtomicU64::new(n) }
    }

    pub fn next(&self) -> u64 {
        self.next.fetch_add(1, Ordering::Relaxed)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashSet;
    use std::sync::Mutex;
    use std::thread;

    #[test]
    fn ids_are_unique_across_threads() {
        let ids = IdGen::starting_at(0);
        let seen = Mutex::new(HashSet::new());

        thread::scope(|s| {
            for _ in 0..8 {
                s.spawn(|| {
                    for _ in 0..1000 {
                        seen.lock().unwrap().insert(ids.next());
                    }
                });
            }
        });

        assert_eq!(seen.into_inner().unwrap().len(), 8000);
    }
}
