use std::sync::Arc;
use std::thread;
use idgen::IdGen;
fn main() {
    let mut collisions = 0;
    for _ in 0..10_000 {
        let ids = Arc::new(IdGen::starting_at(0));
        let other = ids.clone();
        let t = thread::spawn(move || other.next_racy());
        let a = ids.next_racy();
        if a == t.join().unwrap() { collisions += 1; }
    }
    println!("collisions in 10000 runs: {collisions}");
}
