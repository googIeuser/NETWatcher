use std::io::{self, BufRead, Write};

use anyhow::Result;
use netwatcher_core::{
    config,
    models::Config,
    monitor::Engine,
    platform,
    reports,
};
use serde::Deserialize;
use serde_json::{json, Value};

#[derive(Debug, Deserialize)]
struct Request {
    command: String,
    #[serde(default)]
    config: Option<Config>,
    #[serde(default)]
    hours: Option<i64>,
    #[serde(default)]
    days: Option<i64>,
    #[serde(default)]
    path: Option<String>,
}

fn response(data: Value) -> Value {
    json!({"ok": true, "data": data})
}

fn error_response(error: impl ToString) -> Value {
    json!({"ok": false, "error": error.to_string()})
}

fn main() -> Result<()> {
    let initial = config::load().unwrap_or_default();
    let engine = Engine::new(initial);
    let store = engine.store();
    let stdin = io::stdin();
    let mut stdout = io::BufWriter::new(io::stdout().lock());

    for line in stdin.lock().lines() {
        let line = match line {
            Ok(value) => value,
            Err(error) => {
                writeln!(stdout, "{}", error_response(error))?;
                stdout.flush()?;
                continue;
            }
        };
        let request: Request = match serde_json::from_str(&line) {
            Ok(value) => value,
            Err(error) => {
                writeln!(stdout, "{}", error_response(error))?;
                stdout.flush()?;
                continue;
            }
        };
        let result = match request.command.as_str() {
            "hello" => json!({
                "ok": true,
                "data": {
                    "name": "NetWatcher Rust Core",
                    "version": env!("CARGO_PKG_VERSION")
                }
            }),
            "load_settings" => match config::load() {
                Ok(value) => response(serde_json::to_value(value)?),
                Err(error) => error_response(error),
            },
            "save_settings" => match request.config {
                Some(value) => match config::save(&value) {
                    Ok(saved) => {
                        engine.update_config(saved.clone());
                        response(serde_json::to_value(saved)?)
                    }
                    Err(error) => error_response(error),
                },
                None => error_response("config is required"),
            },
            "start" => response(serde_json::to_value(engine.start())?),
            "stop" => response(serde_json::to_value(engine.stop())?),
            "snapshot" => response(serde_json::to_value(engine.snapshot())?),
            "monitor_once" => response(serde_json::to_value(engine.monitor_once())?),
            "get_outages" => match engine.outage_history(request.days.unwrap_or(30)) {
                Ok(items) => response(serde_json::to_value(items)?),
                Err(error) => error_response(error),
            },
            "clear_outages" => {
                match engine.clear_outage_history(request.days.unwrap_or(30)) {
                    Ok(items) => response(serde_json::to_value(items)?),
                    Err(error) => error_response(error),
                }
            },
            "generate_html_report" => match reports::generate_html(&store, request.hours.unwrap_or(24)) {
                Ok(mut report) => {
                    if let Err(error) = platform::open_file(&report.path) {
                        report.message = format!(
                            "HTML report created. Automatic opening failed: {error}"
                        );
                    }
                    response(serde_json::to_value(report)?)
                }
                Err(error) => error_response(error),
            },
            "generate_evidence_report" => match reports::generate_evidence(&store, request.days.unwrap_or(7)) {
                Ok(mut report) => {
                    if let Err(error) = platform::open_file(&report.path) {
                        report.message = format!(
                            "ISP Evidence Report created. Automatic opening failed: {error}"
                        );
                    }
                    response(serde_json::to_value(report)?)
                }
                Err(error) => error_response(error),
            },
            "export_diagnostics" => match reports::export_diagnostics(
                &store,
                &engine.config(),
                &engine.snapshot(),
                request.hours.unwrap_or(168),
            ) {
                Ok(mut report) => {
                    if let Err(error) = platform::open_folder(store.reports_dir()) {
                        report.message = format!(
                            "Diagnostics ZIP created. Reports folder could not be opened: {error}"
                        );
                    }
                    response(serde_json::to_value(report)?)
                }
                Err(error) => error_response(error),
            },
            "open_file" => match request.path {
                Some(path) => match platform::open_file(path) {
                    Ok(()) => response(json!({"opened": true})),
                    Err(error) => error_response(error),
                },
                None => error_response("path is required"),
            },
            "open_reports_folder" => match platform::open_folder(store.reports_dir()) {
                Ok(()) => response(json!({"opened": true})),
                Err(error) => error_response(error),
            },
            "open_logs_folder" => match platform::open_folder(store.dir()) {
                Ok(()) => response(json!({"opened": true})),
                Err(error) => error_response(error),
            },
            "shutdown" => {
                let data = response(json!({"stopped": true}));
                writeln!(stdout, "{data}")?;
                stdout.flush()?;
                let _ = engine.stop();
                break;
            }
            other => error_response(format!("unknown command: {other}")),
        };
        writeln!(stdout, "{result}")?;
        stdout.flush()?;
    }
    Ok(())
}
