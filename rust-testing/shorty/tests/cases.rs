use std::path::Path;

fn decode_case(path: &Path, input: String) -> datatest_stable::Result<()> {
    let (code, expected) = input.trim().split_once(" => ").ok_or("bad case file")?;
    let got = shorty::code::decode(code).map_or("none".to_string(), |id| id.to_string());
    if got != expected {
        return Err(format!("{}: decode({code}) = {got}, want {expected}", path.display()).into());
    }
    Ok(())
}

datatest_stable::harness! {
    { test = decode_case, root = "tests/cases", pattern = r"\.case$" },
}
