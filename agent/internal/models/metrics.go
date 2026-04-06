package models

// AppMetrics represents a snapshot of an application's resource usage
type AppMetrics struct {
	AppID      string  `json:"app_id"`
	CPU        float64 `json:"cpu"`
	Memory     int     `json:"memory"` // In MB
	Status     string  `json:"status"` // running, stopped, etc.
	Timestamp  int64   `json:"timestamp"`
}

// AppConfig represents a monitored application
type AppConfig struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // systemd or docker
	Name     string `json:"name"` // Unit name or container name
	PID      int    `json:"pid"`
}

// ControlCommand represents an incoming command from the UI
type ControlCommand struct {
	Command  string `json:"command"` // START_SERVICE, STOP_SERVICE, RESTART_SERVICE
	AppID    string `json:"app_id"`
}
