use sysinfo::{ProcessRefreshKind, RefreshKind, System, UpdateKind};
use crate::models::AppMetrics;
use std::time::{SystemTime, UNIX_EPOCH};

pub struct Discoverer {
    sys: System,
}

impl Discoverer {
    pub fn new() -> Self {
        Self {
            sys: System::new_with_specifics(
                RefreshKind::new().with_processes(ProcessRefreshKind::new().with_cpu().with_memory()),
            ),
        }
    }

    pub fn discover_and_collect(&mut self) -> Vec<AppMetrics> {
        self.sys.refresh_processes_specifics(ProcessRefreshKind::new().with_cpu().with_memory().with_exe(UpdateKind::OnlyIfNotSet));
        
        let mut metrics = Vec::new();
        let current_ts = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() as i64;
        
        let targets = ["docker", "dockerd", "containerd", "systemd", "nginx", "postgres"];
        
        for (_pid, process) in self.sys.processes() {
            let name = process.name();
            // Just filter the basic processes for this simulation
            if targets.iter().any(|&t| name.contains(t)) {
                metrics.push(AppMetrics {
                    app_id: name.to_string(), // In real app, we'd use robust IDs
                    cpu: process.cpu_usage() as f64,
                    memory: (process.memory() / 1024 / 1024) as i64, // Convert to MB
                    status: "running".to_string(),
                    timestamp: current_ts,
                });
            }
        }
        
        metrics
    }
}
