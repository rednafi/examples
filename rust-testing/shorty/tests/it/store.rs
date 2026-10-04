use shorty::store::MemStore;

#[test]
fn mem_store_passes_the_conformance_suite() {
    shorty::testing::store_conformance(&MemStore::default());
}
