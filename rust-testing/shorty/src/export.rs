use std::io::{self, Write};

/// Writes `code,url` lines.
pub fn export(links: &[(&str, &str)], out: &mut impl Write) -> io::Result<()> {
    for (code, url) in links {
        writeln!(out, "{code},{url}")?;
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn writes_one_line_per_link() {
        let mut buf = Vec::new();
        export(&[("g8", "https://rust-lang.org")], &mut buf).unwrap();
        assert_eq!(buf, b"g8,https://rust-lang.org\n");
    }
}
