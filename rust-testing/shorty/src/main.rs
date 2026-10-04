use std::process::ExitCode;
use std::sync::Arc;

use shorty::code::{decode, encode};
use shorty::config::Config;
use shorty::{http::router, service::Shortener, store::MemStore};

#[tokio::main]
async fn main() -> ExitCode {
    let args: Vec<String> = std::env::args().skip(1).collect();
    match args.iter().map(String::as_str).collect::<Vec<_>>().as_slice() {
        ["encode", n] => match n.parse() {
            Ok(n) => println!("{}", encode(n)),
            Err(_) => return fail(&format!("not a number: {n}")),
        },
        ["decode", code] => match decode(code) {
            Some(id) => println!("{id}"),
            None => return fail(&format!("invalid code: {code}")),
        },
        ["serve"] => {
            let cfg = Config::from_env();
            let app = router(Arc::new(Shortener::new(MemStore::default())));
            let listener = match tokio::net::TcpListener::bind(&cfg.addr).await {
                Ok(l) => l,
                Err(e) => return fail(&format!("binding {}: {e}", cfg.addr)),
            };
            eprintln!("listening on {}", cfg.addr);
            if let Err(e) = axum::serve(listener, app).await {
                return fail(&e.to_string());
            }
        }
        _ => return fail("usage: shorty <encode N | decode CODE | serve>"),
    }
    ExitCode::SUCCESS
}

fn fail(msg: &str) -> ExitCode {
    eprintln!("error: {msg}");
    ExitCode::FAILURE
}
