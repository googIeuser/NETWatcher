use std::{
    env,
    fs,
    path::{Path, PathBuf},
};

use anyhow::{Context, Result};

use crate::models::Config;

pub fn config_dir() -> PathBuf {
    if cfg!(windows) {
        if let Some(app_data) = env::var_os("APPDATA") {
            return PathBuf::from(app_data).join("NetWatcher");
        }
    }
    if let Some(config_home) = env::var_os("XDG_CONFIG_HOME") {
        return PathBuf::from(config_home).join("NetWatcher");
    }
    if let Some(home) = env::var_os("HOME") {
        return PathBuf::from(home).join(".config").join("NetWatcher");
    }
    PathBuf::from(".").join("NetWatcher")
}

pub fn config_path() -> PathBuf {
    config_dir().join("settings.json")
}

pub fn load() -> Result<Config> {
    load_from(&config_path())
}

pub fn load_from(path: &Path) -> Result<Config> {
    if !path.exists() {
        return Ok(Config::default());
    }
    let data = fs::read_to_string(path)
        .with_context(|| format!("failed to read settings from {}", path.display()))?;
    let config: Config = serde_json::from_str(&data)
        .with_context(|| format!("invalid settings JSON at {}", path.display()))?;
    Ok(config.normalised())
}

pub fn save(config: &Config) -> Result<Config> {
    let dir = config_dir();
    fs::create_dir_all(&dir)
        .with_context(|| format!("failed to create {}", dir.display()))?;
    let normalised = config.clone().normalised();
    let data = serde_json::to_string_pretty(&normalised)?;
    fs::write(config_path(), data)
        .with_context(|| "failed to write NetWatcher settings")?;
    Ok(normalised)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn normalises_legacy_values() {
        let config = Config {
            language: "xx".into(),
            theme: "blue".into(),
            interval_seconds: 0.1,
            timeout_ms: 5,
            ..Config::default()
        }
        .normalised();
        assert_eq!(config.language, "en");
        assert_eq!(config.theme, "dark");
        assert_eq!(config.interval_seconds, 2.0);
        assert_eq!(config.timeout_ms, 1500);
    }
}
