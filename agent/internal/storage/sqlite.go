package storage

import (
	"database/sql"
	"log"
	"time"

	"sysagent/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	store := &Store{db: db}
	if err := store.initSchema(); err != nil {
		return nil, err
	}

	// Optional: start a background go-routine to cleanup 24h old data
	go store.cleanupRoutine()

	return store, nil
}

func (s *Store) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS app_metrics (
		app_id TEXT,
		timestamp DATETIME,
		cpu_usage FLOAT,
		ram_usage_mb INT,
		status TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_app_ts ON app_metrics(app_id, timestamp);
	`
	_, err := s.db.Exec(query)
	return err
}

// BatchInsert inserts a batch of metrics efficiently
func (s *Store) BatchInsert(metrics []models.AppMetrics) error {
	if len(metrics) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO app_metrics (app_id, timestamp, cpu_usage, ram_usage_mb, status) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, m := range metrics {
		_, err = stmt.Exec(m.AppID, time.Unix(m.Timestamp, 0), m.CPU, m.Memory, m.Status)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		cutoff := time.Now().Add(-24 * time.Hour)
		_, err := s.db.Exec("DELETE FROM app_metrics WHERE timestamp < ?", cutoff)
		if err != nil {
			log.Println("Error cleaning up old metrics:", err)
		}
	}
}
