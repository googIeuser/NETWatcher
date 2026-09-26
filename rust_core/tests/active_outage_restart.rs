use std::{
    env, fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
    time::{SystemTime, UNIX_EPOCH},
};

use chrono::{Duration, Utc};
use serde_json::{json, Value};

fn get_outages_from_restarted_core(home: &Path) -> Value {
    let mut child = Command::new(env!("CARGO_BIN_EXE_netwatcher_core"))
        .env("USERPROFILE", home)
        .env("APPDATA", home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .unwrap();
    {
        let mut stdin = child.stdin.take().unwrap();
        writeln!(stdin, "{}", json!({"command": "get_outages", "days": 30})).unwrap();
        writeln!(stdin, "{}", json!({"command": "shutdown"})).unwrap();
    }
    let output = child.wait_with_output().unwrap();
    assert!(output.status.success());
    let lines = String::from_utf8(output.stdout).unwrap();
    serde_json::from_str(lines.lines().next().unwrap()).unwrap()
}

#[test]
fn core_restores_an_active_outage_after_restart() {
    let unique = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_nanos();
    let home = env::temp_dir().join(format!(
        "netwatcher-restart-{}-{unique}",
        std::process::id()
    ));
    let logs = home.join("Documents").join("NetWatcherLogs");
    fs::create_dir_all(&logs).unwrap();
    let start = (Utc::now() - Duration::minutes(5)).to_rfc3339();
    fs::write(
        logs.join("active_outage.json"),
        json!({
            "start": start,
            "end": "",
            "category": "offline",
            "details": "Connection lost before restart.",
            "durationSeconds": 0.0,
            "active": true
        })
        .to_string(),
    )
    .unwrap();

    let response = get_outages_from_restarted_core(&home);
    let _ = fs::remove_dir_all(home);

    assert_eq!(response["ok"], true);
    assert_eq!(response["data"][0]["start"], start);
    assert_eq!(response["data"][0]["active"], true);
    assert!(response["data"][0]["durationSeconds"].as_f64().unwrap() > 280.0);
}

#[test]
fn completed_outage_does_not_reappear_as_active_after_restart() {
    let unique = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_nanos();
    let home = env::temp_dir().join(format!(
        "netwatcher-completed-{}-{unique}",
        std::process::id()
    ));
    let logs = home.join("Documents").join("NetWatcherLogs");
    fs::create_dir_all(&logs).unwrap();
    let start = (Utc::now() - Duration::minutes(5)).to_rfc3339();
    let end = Utc::now().to_rfc3339();
    fs::write(
        logs.join("active_outage.json"),
        json!({
            "start": start,
            "end": "",
            "category": "offline",
            "details": "Connection lost before restart.",
            "durationSeconds": 0.0,
            "active": true
        })
        .to_string(),
    )
    .unwrap();
    fs::write(
        logs.join("outages_v4.csv"),
        format!(
            "start,end,category,details,duration_seconds\n{start},{end},offline,Connection recovered,300\n"
        ),
    )
    .unwrap();

    let response = get_outages_from_restarted_core(&home);
    let _ = fs::remove_dir_all(home);

    assert_eq!(response["ok"], true);
    assert_eq!(response["data"].as_array().unwrap().len(), 1);
    assert_eq!(response["data"][0]["active"], false);
}
