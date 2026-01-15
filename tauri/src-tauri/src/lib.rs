use std::process::{Command, Stdio};
use tauri::Manager;

#[tauri::command]
async fn get_app_version() -> String {
    env!("CARGO_PKG_VERSION").to_string()
}

#[tauri::command]
async fn run_cli_command(cmd: String, args: Vec<String>) -> Result<String, String> {
    let output = Command::new(cmd)
        .args(args)
        .output()
        .map_err(|e| e.to_string())?;

    let stdout = String::from_utf8_lossy(&output.stdout).to_string();
    let stderr = String::from_utf8_lossy(&output.stderr).to_string();

    if !output.status.success() {
        return Err(stderr);
    }

    Ok(stdout)
}

#[tauri::command]
async fn check_system_health() -> Result<serde_json::Value, String> {
    let health = serde_json::json!({
        "os": std::env::consts::OS,
        "arch": std::env::consts::ARCH,
        "version": env!("CARGO_PKG_VERSION"),
    });

    Ok(health)
}

#[tauri::command]
async fn open_url(url: String) -> Result<(), String> {
    open::that(&url).map_err(|e| e.to_string())?;
    Ok(())
}

#[tauri::command]
async fn greet(name: String) -> String {
    format!("Hello, {}!", name)
}

// Start the daemon service as a background process
fn start_daemon() -> Option<std::process::Child> {
    // Find the ai-mgr binary - look in common locations
    let binary_paths = vec![
        "/Users/sutang/01_sutang/02_project/ai-manager/ai-mgr",
        "./ai-mgr",
        "ai-mgr",
    ];

    for path in binary_paths {
        if std::path::Path::new(path).exists() {
            println!("Starting daemon from: {}", path);
            match Command::new(path)
                .arg("daemon")
                .stdout(Stdio::null())
                .stderr(Stdio::null())
                .spawn()
            {
                Ok(child) => {
                    println!("Daemon started successfully");
                    return Some(child);
                }
                Err(e) => {
                    println!("Failed to start daemon: {}", e);
                }
            }
            break;
        }
    }

    None
}

pub fn run_app() {
    // Start the daemon service
    let _daemon = start_daemon();

    // Wait a bit for daemon to start
    std::thread::sleep(std::time::Duration::from_secs(1));

    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![
            get_app_version,
            run_cli_command,
            check_system_health,
            open_url,
            greet,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
