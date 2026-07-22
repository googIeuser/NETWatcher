use std::{path::Path, process::Command};

use anyhow::{bail, Result};

fn launch(path: &Path) -> Result<()> {
    #[cfg(windows)]
    {
        let status = Command::new("cmd")
            .args(["/C", "start", ""])
            .arg(path)
            .status()?;
        if !status.success() {
            bail!("Windows could not open {}", path.display());
        }
        return Ok(());
    }
    #[cfg(target_os = "macos")]
    {
        let status = Command::new("open").arg(path).status()?;
        if !status.success() {
            bail!("could not open {}", path.display());
        }
        return Ok(());
    }
    #[cfg(all(unix, not(target_os = "macos")))]
    {
        let status = Command::new("xdg-open").arg(path).status()?;
        if !status.success() {
            bail!("could not open {}", path.display());
        }
        return Ok(());
    }
}

pub fn open_file(path: impl AsRef<Path>) -> Result<()> {
    let path = path.as_ref();
    if !path.exists() {
        bail!("file does not exist: {}", path.display());
    }
    launch(path)
}

pub fn open_folder(path: impl AsRef<Path>) -> Result<()> {
    let path = path.as_ref();
    std::fs::create_dir_all(path)?;
    launch(path)
}
