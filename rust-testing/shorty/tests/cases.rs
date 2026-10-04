use std::path::Path;

use shorty::code::decode;

fn decode_case(path: &Path, input: String) -> datatest_stable::Result<()> {
    let (code, want) = input.trim().split_once(" => ").ok_or("bad case")?;
    let got = decode(code).map_or("none".into(), |id| id.to_string());
    if got != want {
        return Err(format!("{}: got {got}, want {want}", path.display()).into());
    }
    Ok(())
}

datatest_stable::harness! {
    { test = decode_case, root = "tests/cases", pattern = r"\.case$" },
}
