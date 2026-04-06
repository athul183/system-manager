mod models;
mod storage;
mod discovery;
mod server;

use std::time::Duration;
use tokio::sync::broadcast;

#[tokio::main]
async fn main() {
    println!("Starting Rust sysagent...");

    // 1. Initialize SQLite Store
    let store = storage::Store::new("metrics.db").expect("Failed to open DB");
    store.cleanup_routine();

    // 2. Setup Broadcast channel for metrics
    let (tx, _) = broadcast::channel(100);
    
    // 3. Start WS Server
    let tx_server = tx.clone();
    tokio::spawn(async move {
        server::start_server(tx_server).await;
    });

    // 4. Discovery & Collection Loop
    let mut discoverer = discovery::Discoverer::new();
    let mut interval = tokio::time::interval(Duration::from_secs(2));

    loop {
        interval.tick().await;
        
        // Collect metrics
        let metrics = discoverer.discover_and_collect();
        if !metrics.is_empty() {
            // Push to WebSocket clients
            let _ = tx.send(metrics.clone());
            
            // Push to SQLite
            if let Err(e) = store.batch_insert(&metrics) {
                eprintln!("DB Insert error: {}", e);
            }
        }
    }
}
