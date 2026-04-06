use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct AppMetrics {
    #[serde(rename = "app_id")]
    pub app_id: String,
    pub cpu: f64,
    pub memory: i64, // In MB
    pub status: String,
    pub timestamp: i64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ControlCommand {
    pub command: String,
    #[serde(rename = "app_id")]
    pub app_id: String,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ServerPayload {
    #[serde(rename = "type")]
    pub msg_type: String, // "metrics"
    pub payload: Vec<AppMetrics>
}
