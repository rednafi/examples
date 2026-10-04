use std::sync::Arc;

use axum::Router;
use axum::extract::{Path, State};
use axum::http::{StatusCode, header};
use axum::response::{IntoResponse, Response};
use axum::routing::{get, post};

use crate::service::Shortener;
use crate::store::Store;

pub fn router<S: Store + 'static>(svc: Arc<Shortener<S>>) -> Router {
    Router::new()
        .route("/links", post(create::<S>))
        .route("/{code}", get(redirect::<S>))
        .with_state(svc)
}

/// `POST /links` with the long URL as the body. Responds with the code.
async fn create<S: Store>(State(svc): State<Arc<Shortener<S>>>, url: String) -> Response {
    if !url.starts_with("http://") && !url.starts_with("https://") {
        return (StatusCode::BAD_REQUEST, "expected an http(s) URL").into_response();
    }
    let code = svc.shorten(&url);
    (StatusCode::CREATED, code).into_response()
}

/// `GET /{code}` redirects to the stored URL.
async fn redirect<S: Store>(
    State(svc): State<Arc<Shortener<S>>>,
    Path(code): Path<String>,
) -> Response {
    match svc.resolve(&code) {
        Some(url) => (StatusCode::TEMPORARY_REDIRECT, [(header::LOCATION, url)]).into_response(),
        None => StatusCode::NOT_FOUND.into_response(),
    }
}
