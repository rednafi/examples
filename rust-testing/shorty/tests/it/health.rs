use shorty::health::is_alive;
use wiremock::matchers::{method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

#[tokio::test]
async fn live_target_is_alive() {
    let server = MockServer::start().await;
    Mock::given(method("HEAD"))
        .and(path("/page"))
        .respond_with(ResponseTemplate::new(200))
        .expect(1)
        .mount(&server)
        .await;

    let url = format!("{}/page", server.uri());
    assert!(is_alive(&reqwest::Client::new(), &url).await);
}
