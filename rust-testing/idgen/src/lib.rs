#[cfg(loom)]
use loom::sync::atomic::{AtomicU64, Ordering};
#[cfg(not(loom))]
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

    /// Deliberately broken: read and write are two separate steps.
    pub fn next_racy(&self) -> u64 {
        let n = self.next.load(Ordering::Relaxed);
        self.next.store(n + 1, Ordering::Relaxed);
        n
    }
}

#[cfg(all(test, loom))]
mod loom_tests {
    use super::*;
    use loom::{sync::Arc, thread};

    #[test]
    #[ignore = "fails on purpose: run with --ignored to watch loom catch it"]
    fn racy_ids_are_unique() {
        loom::model(|| {
            let ids = Arc::new(IdGen::starting_at(0));
            let other = ids.clone();
            let t = thread::spawn(move || other.next_racy());
            let a = ids.next_racy();
            assert_ne!(a, t.join().unwrap());
        });
    }

    #[test]
    fn ids_are_unique() {
        loom::model(|| {
            let ids = Arc::new(IdGen::starting_at(0));
            let other = ids.clone();
            let t = thread::spawn(move || other.next());
            let a = ids.next();
            assert_ne!(a, t.join().unwrap());
        });
    }
}

#[cfg(all(test, not(loom)))]
mod tests {
    use super::*;
    use std::collections::HashSet;
    use std::thread;

    #[test]
    fn ids_are_unique_across_threads() {
        let ids = IdGen::starting_at(0);
        let all: Vec<u64> = thread::scope(|s| {
            let handles: Vec<_> = (0..8)
                .map(|_| s.spawn(|| (0..1000).map(|_| ids.next()).collect::<Vec<_>>()))
                .collect();
            handles.into_iter().flat_map(|h| h.join().unwrap()).collect()
        });
        let unique: HashSet<_> = all.iter().collect();
        assert_eq!(unique.len(), 8000);
    }
}
