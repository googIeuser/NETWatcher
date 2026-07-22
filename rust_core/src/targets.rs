use sha1::{Digest, Sha1};
use url::Url;

use crate::models::Target;

fn target_id(value: &str) -> String {
    let mut hasher = Sha1::new();
    hasher.update(value.trim().to_ascii_lowercase().as_bytes());
    let output = hasher.finalize();
    output[..6].iter().map(|byte| format!("{byte:02x}")).collect()
}

pub fn default_targets() -> Vec<Target> {
    let mut targets = Vec::new();
    if let Some(gateway) = detect_default_gateway() {
        targets.push(Target {
            id: target_id(&format!("gateway:{gateway}")),
            name: "Default Gateway".into(),
            host: gateway,
            kind: "local".into(),
            mode: "ping".into(),
            custom: false,
        });
    }
    targets.extend([
        Target {
            id: target_id("cloudflare:1.1.1.1"),
            name: "Cloudflare".into(),
            host: "1.1.1.1".into(),
            kind: "internet".into(),
            mode: "ping".into(),
            custom: false,
        },
        Target {
            id: target_id("google:8.8.8.8"),
            name: "Google".into(),
            host: "8.8.8.8".into(),
            kind: "internet".into(),
            mode: "ping".into(),
            custom: false,
        },
    ]);
    targets
}

pub fn parse_target(raw: &str) -> Target {
    let raw = raw.trim();
    let mut target = Target {
        id: target_id(raw),
        name: format!("Custom: {raw}"),
        host: raw.into(),
        kind: "internet".into(),
        mode: "ping".into(),
        custom: true,
    };
    let lower = raw.to_ascii_lowercase();
    if lower.starts_with("tcp://") {
        let host = raw[6..].trim();
        target.host = host.into();
        target.name = format!("TCP: {host}");
        target.mode = "tcp".into();
    } else if lower.starts_with("http://") || lower.starts_with("https://") {
        if let Ok(url) = Url::parse(raw) {
            if let Some(host) = url.host_str() {
                target.name = format!("{}: {host}", url.scheme().to_ascii_uppercase());
                target.mode = url.scheme().to_ascii_lowercase();
            }
        }
    }
    target
}

pub fn all_targets(custom: &[String]) -> Vec<Target> {
    let mut values = default_targets();
    for raw in custom {
        let target = parse_target(raw);
        if target.host.trim().is_empty()
            || values.iter().any(|existing| existing.id == target.id)
        {
            continue;
        }
        values.push(target);
    }
    values
}

fn detect_default_gateway() -> Option<String> {
    #[cfg(windows)]
    {
        use std::process::Command;
        let output = Command::new("route")
            .args(["print", "-4", "0.0.0.0"])
            .output()
            .ok()?;
        let text = String::from_utf8_lossy(&output.stdout);
        for line in text.lines() {
            let columns: Vec<_> = line.split_whitespace().collect();
            if columns.len() >= 4 && columns[0] == "0.0.0.0" && columns[1] == "0.0.0.0" {
                return Some(columns[2].to_string());
            }
        }
    }
    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parses_supported_custom_target_formats() {
        assert_eq!(parse_target("1.1.1.1").mode, "ping");
        assert_eq!(parse_target("tcp://example.com:443").mode, "tcp");
        assert_eq!(parse_target("https://example.com/test").mode, "https");
    }
}
