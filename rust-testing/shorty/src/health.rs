use std::time::Duration;

/// Calls `probe` up to three times, sleeping 1s, 2s, 4s after failures.
pub async fn with_backoff<F, Fut>(mut probe: F) -> bool
where
    F: FnMut() -> Fut,
    Fut: Future<Output = bool>,
{
    for attempt in 0..3 {
        if probe().await {
            return true;
        }
        tokio::time::sleep(Duration::from_secs(1 << attempt)).await;
    }
    false
}

/// Checks that a target URL still answers.
pub async fn is_alive(client: &reqwest::Client, url: &str) -> bool {
    with_backoff(|| async {
        matches!(client.head(url).send().await, Ok(r) if r.status().is_success())
    })
    .await
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::time::Instant;

    #[tokio::test(start_paused = true)]
    async fn gives_up_after_three_tries() {
        let mut calls = 0;
        let start = Instant::now();

        let ok = with_backoff(|| {
            calls += 1;
            async { false }
        })
        .await;

        assert!(!ok);
        assert_eq!(calls, 3);
        assert_eq!(start.elapsed(), Duration::from_secs(7));
    }
}
