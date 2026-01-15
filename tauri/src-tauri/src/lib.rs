use std::process::{Command, Stdio};
use std::sync::Arc;
use std::io::{Read, Write};
use std::net::TcpListener;

#[tauri::command]
async fn get_app_version() -> String {
    env!("CARGO_PKG_VERSION").to_string()
}

// Simple HTTP server to serve frontend files + API proxy
fn start_http_server(port: u16, resources_path: String) {
    use std::time::Duration;

    let listener = TcpListener::bind(("127.0.0.1", port)).expect("Failed to bind to port");
    let resources_path = Arc::new(resources_path);

    std::thread::spawn(move || {
        for stream in listener.incoming() {
            if let Ok(mut stream) = stream {
                // Set timeout to avoid blocking forever
                let _ = stream.set_read_timeout(Some(Duration::from_secs(5)));
                let _ = stream.set_write_timeout(Some(Duration::from_secs(5)));

                // Read request with timeout
                let mut request = [0; 8192];
                let mut bytes_read = 0;
                let deadline = std::time::Instant::now() + Duration::from_secs(5);

                loop {
                    if deadline.elapsed() > Duration::from_secs(5) {
                        break;
                    }
                    match stream.read(&mut request[bytes_read..]) {
                        Ok(0) => break,
                        Ok(n) => {
                            bytes_read += n;
                            // Check for end of HTTP headers
                            if bytes_read >= 4 {
                                for i in 0..bytes_read.saturating_sub(3) {
                                    if &request[i..i+4] == b"\r\n\r\n" {
                                        bytes_read = i + 4;
                                        break;
                                    }
                                }
                                if bytes_read < 8192 && bytes_read >= 4 {
                                    // Check if we found the headers end
                                    let mut found = false;
                                    for i in 0..bytes_read.saturating_sub(3) {
                                        if &request[i..i+4] == b"\r\n\r\n" {
                                            bytes_read = i + 4;
                                            found = true;
                                            break;
                                        }
                                    }
                                    if found {
                                        break;
                                    }
                                }
                            }
                            if bytes_read >= request.len() {
                                break;
                            }
                        }
                        Err(ref e) if e.kind() == std::io::ErrorKind::WouldBlock => break,
                        Err(_) => break,
                    }
                }

                if bytes_read == 0 {
                    continue;
                }

                let request_str = String::from_utf8_lossy(&request[..bytes_read]);
                let lines: Vec<&str> = request_str.lines().collect();

                // Parse path
                let path = lines.first()
                    .and_then(|l| l.split_whitespace().nth(1))
                    .unwrap_or("/")
                    .to_string();

                // API proxy
                if path.starts_with("/api/") {
                    if let Ok(mut daemon_conn) = std::net::TcpStream::connect("127.0.0.1:19999") {
                        daemon_conn.set_read_timeout(Some(Duration::from_secs(5))).ok();
                        daemon_conn.set_write_timeout(Some(Duration::from_secs(5))).ok();

                        let modified = request_str
                            .replace("Host: 127.0.0.1:3456", "Host: 127.0.0.1:19999")
                            .replace("Host: localhost:3456", "Host: 127.0.0.1:19999");

                        if daemon_conn.write_all(modified.as_bytes()).is_ok() {
                            let mut response = Vec::new();
                            let mut buf = [0; 8192];
                            let resp_deadline = std::time::Instant::now() + Duration::from_secs(5);

                            while resp_deadline.elapsed() < Duration::from_secs(5) {
                                match daemon_conn.read(&mut buf) {
                                    Ok(0) => break,
                                    Ok(n) => {
                                        response.extend_from_slice(&buf[..n]);
                                        // Check if complete
                                        if let Some(pos) = response.windows(4).position(|w| w == b"\r\n\r\n") {
                                            let header_end = pos + 4;
                                            let headers = String::from_utf8_lossy(&response[..header_end]);
                                            if let Some(cl) = headers.find("Content-Length:").and_then(|i| {
                                                headers[i..].lines().next()?.split_whitespace().nth(1)?.parse::<usize>().ok()
                                            }) {
                                                if response.len() >= header_end + cl {
                                                    break;
                                                }
                                            }
                                        }
                                        if response.len() > 100000 {
                                            break;
                                        }
                                    }
                                    Err(_) => break,
                                }
                            }
                            let _ = stream.write_all(&response);
                        }
                    } else {
                        let _ = stream.write_all(b"HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n");
                    }
                    continue;
                }

                // Serve static files
                let file_path = if path == "/" || path.is_empty() {
                    format!("{}/index.html", resources_path.as_str())
                } else {
                    format!("{}{}", resources_path.as_str(), path)
                };

                if let Ok(mut file) = std::fs::File::open(&file_path) {
                    let mut contents = Vec::new();
                    if file.read_to_end(&mut contents).is_ok() {
                        let ct = if file_path.ends_with(".html") { "text/html" }
                            else if file_path.ends_with(".js") { "application/javascript" }
                            else if file_path.ends_with(".css") { "text/css" }
                            else if file_path.ends_with(".svg") { "image/svg+xml" }
                            else if file_path.ends_with(".png") { "image/png" }
                            else if file_path.ends_with(".ico") { "image/x-icon" }
                            else { "application/octet-stream" };

                        let header = format!("HTTP/1.1 200 OK\r\nContent-Type: {}\r\nContent-Length: {}\r\n\r\n", ct, contents.len());
                        let _ = stream.write_all(header.as_bytes());
                        let _ = stream.write_all(&contents);
                    }
                } else if path == "/" || path.is_empty() {
                    // Fallback
                    let index_path = format!("{}/index.html", resources_path.as_str());
                    if let Ok(mut file) = std::fs::File::open(&index_path) {
                        let mut contents = Vec::new();
                        if file.read_to_end(&mut contents).is_ok() {
                            let header = format!("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: {}\r\n\r\n", contents.len());
                            let _ = stream.write_all(header.as_bytes());
                            let _ = stream.write_all(&contents);
                        }
                    }
                }
            }
        }
    });
}

fn get_resources_path() -> String {
    // Get the resources path from the environment (set by Tauri)
    if let Ok(path) = std::env::var("RESOURCE_PATH") {
        return path;
    }

    // Fallback: try to determine from the executable path
    let exe_path = std::env::current_exe().unwrap_or_default();
    let exe_dir = exe_path.parent().unwrap_or(std::path::Path::new(""));

    // Navigate from Contents/MacOS to Resources/_up_/_up_/ui/dist
    let resources_path = exe_dir
        .parent()
        .unwrap_or(std::path::Path::new(""))
        .join("Resources")
        .join("_up_")
        .join("_up_")
        .join("ui")
        .join("dist");

    resources_path.to_string_lossy().to_string()
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

// Check if daemon is already running
fn is_daemon_running() -> bool {
    std::net::TcpStream::connect("127.0.0.1:19999").is_ok()
}

// Wait for daemon to be ready
fn wait_for_daemon(timeout_ms: u64) -> bool {
    let start = std::time::Instant::now();
    while start.elapsed().as_millis() < timeout_ms as u128 {
        if is_daemon_running() {
            return true;
        }
        std::thread::sleep(std::time::Duration::from_millis(100));
    }
    false
}

// Start the daemon service as a background process
fn start_daemon() -> Option<std::process::Child> {
    // Check if already running
    if is_daemon_running() {
        println!("Daemon already running on port 19999");
        return None;
    }

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
                    // Wait for daemon to be ready
                    if wait_for_daemon(5000) {
                        println!("Daemon is ready");
                    } else {
                        println!("Warning: Daemon may not be ready yet");
                    }
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

    // Start HTTP server for frontend
    let port = 3456;
    let resources_path = get_resources_path();
    println!("Starting HTTP server on port {} for resources: {}", port, resources_path);

    start_http_server(port, resources_path);

    // Wait a bit for server to start
    std::thread::sleep(std::time::Duration::from_millis(500));

    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![
            get_app_version,
            run_cli_command,
            check_system_health,
            open_url,
            greet,
        ])
        .setup(|app| {
            let window = tauri::WebviewWindowBuilder::new(
                app,
                "main",
                tauri::WebviewUrl::App("http://127.0.0.1:3456/".parse().unwrap())
            )
            .title("OpenShrimp")
            .inner_size(1200.0, 800.0)
            .resizable(true)
            .fullscreen(false)
            .build()?;

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
