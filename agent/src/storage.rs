use rusqlite::{params, Connection, Result};
use std::sync::{Arc, Mutex};
use std::time::{SystemTime, UNIX_EPOCH};
use crate::models::AppMetrics;

#[derive(Clone)]
pub struct Store {
    conn: Arc<Mutex<Connection>>,
}

impl Store {
    pub fn new(db_path: &str) -> Result<Self> {
        let conn = Connection::open(db_path)?;
        let store = Self {
            conn: Arc::new(Mutex::new(conn)),
        };
        store.init_schema()?;
        Ok(store)
    }

    fn init_schema(&self) -> Result<()> {
        let conn = self.conn.lock().unwrap();
        conn.execute(
            "CREATE TABLE IF NOT EXISTS app_metrics (
                app_id TEXT,
                timestamp DATETIME,
                cpu_usage FLOAT,
                ram_usage_mb INT,
                status TEXT
            );",
            [],
        )?;
        conn.execute(
            "CREATE INDEX IF NOT EXISTS idx_app_ts ON app_metrics(app_id, timestamp);",
            [],
        )?;
        Ok(())
    }

    pub fn batch_insert(&self, metrics: &[AppMetrics]) -> Result<()> {
        if metrics.is_empty() {
            return Ok(());
        }
        let mut conn = self.conn.lock().unwrap();
        let tx = conn.transaction()?;
        {
            let mut stmt = tx.prepare(
                "INSERT INTO app_metrics (app_id, timestamp, cpu_usage, ram_usage_mb, status) VALUES (?1, ?2, ?3, ?4, ?5)"
            )?;
            for m in metrics {
                stmt.execute(params![m.app_id, m.timestamp, m.cpu, m.memory, m.status])?;
            }
        }
        tx.commit()?;
        Ok(())
    }

    pub fn cleanup_routine(&self) {
        let conn_clone = self.conn.clone();
        tokio::spawn(async move {
            let mut interval = tokio::time::interval(std::time::Duration::from_secs(3600));
            loop {
                interval.tick().await;
                let cutoff = SystemTime::now()
                    .duration_since(UNIX_EPOCH)
                    .unwrap()
                    .as_secs() as i64
                    - 86400; // 24 hours ago
                if let Ok(conn) = conn_clone.lock() {
                    let _ = conn.execute("DELETE FROM app_metrics WHERE timestamp < ?1", params![cutoff]);
                }
            }
        });
    }
}
