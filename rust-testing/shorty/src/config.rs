use std::time::Duration;

#[derive(Debug, Clone, PartialEq)]
pub struct Config {
    pub addr: String,
    pub ttl: Duration,
}

impl Config {
    pub fn from_env() -> Self {
        Self::from_lookup(|key| std::env::var(key).ok())
    }

    pub fn from_lookup(get: impl Fn(&str) -> Option<String>) -> Self {
        let addr = get("SHORTY_ADDR").unwrap_or_else(|| "0.0.0.0:8080".into());
        let ttl = get("SHORTY_TTL_SECS")
            .and_then(|s| s.parse().ok())
            .map_or(Duration::from_secs(86_400), Duration::from_secs);
        Self { addr, ttl }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;

    #[test]
    fn reads_ttl_without_touching_the_process_env() {
        let env = HashMap::from([("SHORTY_TTL_SECS", "60")]);
        let cfg = Config::from_lookup(|k| env.get(k).map(|v| v.to_string()));
        assert_eq!(cfg.ttl, Duration::from_secs(60));
    }
}
