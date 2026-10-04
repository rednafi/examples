use std::sync::Arc;

use axum::body::Body;
use axum::http::{Request, StatusCode, header};
use shorty::{http::router, service::Shortener, store::MemStore};
use http_body_util::BodyExt;
use tower::ServiceExt;

#[tokio::test]
async fn known_code_redirects() {
    let svc = Arc::new(Shortener::new(MemStore::default()));
    let code = svc.shorten("https://rust-lang.org");

    let req = Request::get(format!("/{code}")).body(Body::empty()).unwrap();
    let resp = router(svc).oneshot(req).await.unwrap();

    assert_eq!(resp.status(), StatusCode::TEMPORARY_REDIRECT);
    assert_eq!(resp.headers()[header::LOCATION], "https://rust-lang.org");
}

#[tokio::test]
async fn real_socket_on_a_random_port() {
    let svc = Arc::new(Shortener::new(MemStore::default()));
    let code = svc.shorten("https://crates.io");

    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let addr = listener.local_addr().unwrap();
    tokio::spawn(axum::serve(listener, router(svc)).into_future());

    let client = reqwest::Client::builder()
        .redirect(reqwest::redirect::Policy::none())
        .build()
        .unwrap();
    let resp = client.get(format!("http://{addr}/{code}")).send().await.unwrap();
    assert_eq!(resp.status(), 307);
}

#[tokio::test]
async fn post_then_follow() {
    let app = router(Arc::new(Shortener::new(MemStore::default())));

    let req = Request::post("/links")
        .body(Body::from("https://rust-lang.org"))
        .unwrap();
    let resp = app.clone().oneshot(req).await.unwrap();
    assert_eq!(resp.status(), StatusCode::CREATED);
    let body = resp.into_body().collect().await.unwrap().to_bytes();
    assert_eq!(&body[..], b"g8");

    let req = Request::get("/g8").body(Body::empty()).unwrap();
    let resp = app.oneshot(req).await.unwrap();
    assert_eq!(resp.status(), StatusCode::TEMPORARY_REDIRECT);
    assert_eq!(resp.headers()[header::LOCATION], "https://rust-lang.org");
}

#[tokio::test]
async fn unknown_code_is_404() {
    let app = router(Arc::new(Shortener::new(MemStore::default())));
    let req = Request::get("/nope").body(Body::empty()).unwrap();
    assert_eq!(app.oneshot(req).await.unwrap().status(), StatusCode::NOT_FOUND);
}
