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
