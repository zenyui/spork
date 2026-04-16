package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

func openDB() (*sql.DB, error) {
	home, err := sporkHome()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(home, 0o750); err != nil {
		return nil, fmt.Errorf("creating spork directory: %w", err)
	}
	dbPath := home + "/spork.db"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	if err := migrateDB(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrating database: %w", err)
	}
	return db, nil
}

func migrateDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS spork_tasks (
			spork_path TEXT NOT NULL,
			task_id    TEXT NOT NULL,
			linked_at  TEXT NOT NULL,
			PRIMARY KEY (spork_path, task_id)
		);
	`)
	return err
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// taskIDsForSpork returns all task IDs linked to a spork path.
func taskIDsForSpork(db *sql.DB, sporkPath string) ([]string, error) {
	rows, err := db.Query(
		`SELECT task_id FROM spork_tasks WHERE spork_path = ? ORDER BY linked_at`, sporkPath,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// sporkPathsForTask returns all spork paths linked to a task ID.
func sporkPathsForTask(db *sql.DB, taskID string) ([]string, error) {
	rows, err := db.Query(
		`SELECT spork_path FROM spork_tasks WHERE task_id = ? ORDER BY linked_at`, taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}

func linkSporkTask(db *sql.DB, sporkPath, taskID string) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO spork_tasks (spork_path, task_id, linked_at) VALUES (?, ?, ?)`,
		sporkPath, taskID, now(),
	)
	return err
}

func unlinkSporkTask(db *sql.DB, sporkPath, taskID string) error {
	res, err := db.Exec(
		`DELETE FROM spork_tasks WHERE spork_path = ? AND task_id = ?`,
		sporkPath, taskID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task %q is not linked to this spork", taskID)
	}
	return nil
}

func deleteLinksForTask(db *sql.DB, taskID string) error {
	_, err := db.Exec(`DELETE FROM spork_tasks WHERE task_id = ?`, taskID)
	return err
}

func deleteLinksForSpork(db *sql.DB, sporkPath string) error {
	_, err := db.Exec(`DELETE FROM spork_tasks WHERE spork_path = ?`, sporkPath)
	return err
}
