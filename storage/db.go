package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
	mu   sync.RWMutex
}

var instance *DB
var once sync.Once

func GetDB() (*DB, error) {
	var err error
	once.Do(func() {
		instance, err = initDB()
	})
	return instance, err
}

func initDB() (*DB, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	conn.SetMaxOpenConns(1)

	db := &DB{conn: conn}

	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func getDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".golem", "golem.db"), nil
}

func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS preferences (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS collections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS request_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		method TEXT NOT NULL,
		headers TEXT,
		body TEXT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		response_status TEXT,
		response_body TEXT,
		response_headers TEXT,
		response_time_ms INTEGER,
		response_size INTEGER,
		is_favorite BOOLEAN DEFAULT 0,
		collection_id INTEGER,
		FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS saved_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		method TEXT NOT NULL,
		headers TEXT,
		body TEXT,
		collection_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_request_history_timestamp ON request_history(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_request_history_url ON request_history(url);
	CREATE INDEX IF NOT EXISTS idx_request_history_method ON request_history(method);
	CREATE INDEX IF NOT EXISTS idx_saved_requests_collection ON saved_requests(collection_id);
	`

	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.conn.Exec(query, args...)
}

func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.conn.Query(query, args...)
}

func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.conn.QueryRow(query, args...)
}

func (db *DB) Begin() (*sql.Tx, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.conn.Begin()
}
