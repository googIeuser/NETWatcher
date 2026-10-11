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
        use std::net::Ipv4Addr;
        use std::sync::{Mutex, OnceLock};
        use std::time::{Duration, Instant};
        use windows_sys::Win32::NetworkManagement::IpHelper::{GetBestRoute, MIB_IPFORWARDROW};

        static CACHE: OnceLock<Mutex<Option<(Instant, Option<String>)>>> = OnceLock::new();
        let mut cache = CACHE.get_or_init(|| Mutex::new(None)).lock().ok()?;
        if let Some((checked_at, gateway)) = cache.as_ref() {
            if checked_at.elapsed() < Duration::from_secs(30) {
                return gateway.clone();
            }
        }

        // Ask Windows for the best route to 0.0.0.0, which is the default route,
        // instead of starting `route print`.
        // SAFETY: MIB_IPFORWARDROW is plain data and GetBestRoute only writes into it.
        let mut route: MIB_IPFORWARDROW = unsafe { std::mem::zeroed() };
        let result = unsafe { GetBestRoute(0, 0, &mut route) };
        let next_hop = route.dwForwardNextHop;
        let gateway = (result == 0 && next_hop != 0)
            .then(|| Ipv4Addr::from(next_hop.to_ne_bytes()).to_string());
        *cache = Some((Instant::now(), gateway.clone()));
        return gateway;
    }
    #[cfg(not(windows))]
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
