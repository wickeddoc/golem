package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Preference struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Collection struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type RequestHistory struct {
	ID              int       `json:"id"`
	URL             string    `json:"url"`
	Method          string    `json:"method"`
	Headers         string    `json:"headers,omitempty"`
	Body            string    `json:"body,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
	ResponseStatus  string    `json:"response_status,omitempty"`
	ResponseBody    string    `json:"response_body,omitempty"`
	ResponseHeaders string    `json:"response_headers,omitempty"`
	ResponseTimeMs  int       `json:"response_time_ms"`
	ResponseSize    int       `json:"response_size"`
	IsFavorite      bool      `json:"is_favorite"`
	CollectionID    *int      `json:"collection_id,omitempty"`
}

type SavedRequest struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Method       string    `json:"method"`
	Headers      string    `json:"headers,omitempty"`
	Body         string    `json:"body,omitempty"`
	CollectionID *int      `json:"collection_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func (db *DB) GetPreference(key string) (*Preference, error) {
	var pref Preference
	err := db.QueryRow(
		"SELECT key, value, updated_at FROM preferences WHERE key = ?",
		key,
	).Scan(&pref.Key, &pref.Value, &pref.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pref, err
}

func (db *DB) SetPreference(key, value string) error {
	_, err := db.Exec(
		`INSERT INTO preferences (key, value, updated_at)
		 VALUES (?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(key) DO UPDATE SET
		 value = excluded.value,
		 updated_at = CURRENT_TIMESTAMP`,
		key, value,
	)
	return err
}

func (db *DB) GetAllPreferences() (map[string]string, error) {
	rows, err := db.Query("SELECT key, value FROM preferences")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefs := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		prefs[key] = value
	}
	return prefs, rows.Err()
}

func (db *DB) SaveRequestHistory(req *RequestHistory) error {
	result, err := db.Exec(
		`INSERT INTO request_history (
			url, method, headers, body, timestamp,
			response_status, response_body, response_headers,
			response_time_ms, response_size, is_favorite, collection_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.URL, req.Method, req.Headers, req.Body, req.Timestamp,
		req.ResponseStatus, req.ResponseBody, req.ResponseHeaders,
		req.ResponseTimeMs, req.ResponseSize, req.IsFavorite, req.CollectionID,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		req.ID = int(id)
	}
	return err
}

func (db *DB) GetRequestHistory(limit int, offset int) ([]*RequestHistory, error) {
	query := `
		SELECT id, url, method, headers, body, timestamp,
			   response_status, response_body, response_headers,
			   response_time_ms, response_size, is_favorite, collection_id
		FROM request_history
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*RequestHistory
	for rows.Next() {
		var req RequestHistory
		var collectionID sql.NullInt64

		err := rows.Scan(
			&req.ID, &req.URL, &req.Method, &req.Headers, &req.Body, &req.Timestamp,
			&req.ResponseStatus, &req.ResponseBody, &req.ResponseHeaders,
			&req.ResponseTimeMs, &req.ResponseSize, &req.IsFavorite, &collectionID,
		)
		if err != nil {
			return nil, err
		}

		if collectionID.Valid {
			id := int(collectionID.Int64)
			req.CollectionID = &id
		}

		history = append(history, &req)
	}

	return history, rows.Err()
}

func (db *DB) SearchRequestHistory(searchTerm string, limit int) ([]*RequestHistory, error) {
	query := `
		SELECT id, url, method, headers, body, timestamp,
			   response_status, response_body, response_headers,
			   response_time_ms, response_size, is_favorite, collection_id
		FROM request_history
		WHERE url LIKE ? OR method LIKE ? OR response_status LIKE ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := db.Query(query, searchPattern, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*RequestHistory
	for rows.Next() {
		var req RequestHistory
		var collectionID sql.NullInt64

		err := rows.Scan(
			&req.ID, &req.URL, &req.Method, &req.Headers, &req.Body, &req.Timestamp,
			&req.ResponseStatus, &req.ResponseBody, &req.ResponseHeaders,
			&req.ResponseTimeMs, &req.ResponseSize, &req.IsFavorite, &collectionID,
		)
		if err != nil {
			return nil, err
		}

		if collectionID.Valid {
			id := int(collectionID.Int64)
			req.CollectionID = &id
		}

		history = append(history, &req)
	}

	return history, rows.Err()
}

func (db *DB) DeleteRequestHistory(id int) error {
	_, err := db.Exec("DELETE FROM request_history WHERE id = ?", id)
	return err
}

func (db *DB) ClearRequestHistory() error {
	_, err := db.Exec("DELETE FROM request_history")
	return err
}

func (db *DB) CreateCollection(name, description string) (*Collection, error) {
	result, err := db.Exec(
		"INSERT INTO collections (name, description, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
		name, description,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Collection{
		ID:          int(id),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}, nil
}

func (db *DB) GetCollections() ([]*Collection, error) {
	rows, err := db.Query("SELECT id, name, description, created_at FROM collections ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []*Collection
	for rows.Next() {
		var col Collection
		if err := rows.Scan(&col.ID, &col.Name, &col.Description, &col.CreatedAt); err != nil {
			return nil, err
		}
		collections = append(collections, &col)
	}

	return collections, rows.Err()
}

func (db *DB) DeleteCollection(id int) error {
	_, err := db.Exec("DELETE FROM collections WHERE id = ?", id)
	return err
}

func (db *DB) SaveRequest(req *SavedRequest) error {
	result, err := db.Exec(
		`INSERT INTO saved_requests (
			name, url, method, headers, body, collection_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		req.Name, req.URL, req.Method, req.Headers, req.Body, req.CollectionID,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		req.ID = int(id)
	}
	return err
}

func (db *DB) GetSavedRequests(collectionID *int) ([]*SavedRequest, error) {
	var rows *sql.Rows
	var err error

	if collectionID != nil {
		rows, err = db.Query(
			`SELECT id, name, url, method, headers, body, collection_id, created_at
			 FROM saved_requests WHERE collection_id = ? ORDER BY name`,
			*collectionID,
		)
	} else {
		rows, err = db.Query(
			`SELECT id, name, url, method, headers, body, collection_id, created_at
			 FROM saved_requests WHERE collection_id IS NULL ORDER BY name`,
		)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*SavedRequest
	for rows.Next() {
		var req SavedRequest
		var collID sql.NullInt64

		err := rows.Scan(
			&req.ID, &req.Name, &req.URL, &req.Method,
			&req.Headers, &req.Body, &collID, &req.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if collID.Valid {
			id := int(collID.Int64)
			req.CollectionID = &id
		}

		requests = append(requests, &req)
	}

	return requests, rows.Err()
}

func (db *DB) GetSavedRequest(id int) (*SavedRequest, error) {
	var req SavedRequest
	var collectionID sql.NullInt64

	err := db.QueryRow(
		`SELECT id, name, url, method, headers, body, collection_id, created_at
		 FROM saved_requests WHERE id = ?`,
		id,
	).Scan(
		&req.ID, &req.Name, &req.URL, &req.Method,
		&req.Headers, &req.Body, &collectionID, &req.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("saved request not found")
	}
	if err != nil {
		return nil, err
	}

	if collectionID.Valid {
		id := int(collectionID.Int64)
		req.CollectionID = &id
	}

	return &req, nil
}

func (db *DB) DeleteSavedRequest(id int) error {
	_, err := db.Exec("DELETE FROM saved_requests WHERE id = ?", id)
	return err
}

func (db *DB) ExportHistory(filepath string) error {
	history, err := db.GetRequestHistory(10000, 0)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return writeFile(filepath, data)
}

func (db *DB) ImportHistory(filepath string) error {
	data, err := readFile(filepath)
	if err != nil {
		return err
	}

	var history []*RequestHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, req := range history {
		_, err := tx.Exec(
			`INSERT INTO request_history (
				url, method, headers, body, timestamp,
				response_status, response_body, response_headers,
				response_time_ms, response_size, is_favorite, collection_id
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			req.URL, req.Method, req.Headers, req.Body, req.Timestamp,
			req.ResponseStatus, req.ResponseBody, req.ResponseHeaders,
			req.ResponseTimeMs, req.ResponseSize, req.IsFavorite, req.CollectionID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func writeFile(filepath string, data []byte) error {
	return os.WriteFile(filepath, data, 0644)
}

func readFile(filepath string) ([]byte, error) {
	return os.ReadFile(filepath)
}
