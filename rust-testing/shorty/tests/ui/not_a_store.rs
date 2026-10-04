use shorty::service::Shortener;

fn main() {
    let svc = Shortener::new(42_u32);
    svc.shorten("https://rust-lang.org");
}
