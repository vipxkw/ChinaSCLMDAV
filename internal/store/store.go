package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("store: not found")

// Store wraps the SQLite database and the data directory.
type Store struct {
	db      *sql.DB
	dataDir string
}

// Config holds the parameters needed to open a store.
type Config struct {
	// DataDir is the root directory served by WebDAV.
	DataDir string
	// DBPath is the path to the SQLite database file.
	DBPath string
}

// Open opens (or creates) the database and creates the data directory layout.
func Open(cfg Config) (*Store, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, err
	}
	// Keep concurrency sane for the single-server workload.
	db.SetMaxOpenConns(1)
	s := &Store{
		db:      db,
		dataDir: filepath.Clean(cfg.DataDir),
	}
	if err := s.init(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) init() error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL DEFAULT '',
	display_name TEXT NOT NULL DEFAULT '',
	password_hash TEXT NOT NULL,
	totp_secret TEXT NOT NULL DEFAULT '',
	totp_enabled INTEGER NOT NULL DEFAULT 0,
	totp_forced INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
	token TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	created_at DATETIME NOT NULL,
	expires_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS app_passwords (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	app_name TEXT NOT NULL,
	secret TEXT NOT NULL DEFAULT '',
	hash TEXT NOT NULL,
	created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS shares (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	token TEXT NOT NULL UNIQUE,
	path TEXT NOT NULL,
	name TEXT NOT NULL,
	size INTEGER NOT NULL DEFAULT 0,
	expires_at DATETIME,
	download_count INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS trash (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	original_path TEXT NOT NULL,
	name TEXT NOT NULL,
	trash_path TEXT NOT NULL,
	is_dir INTEGER NOT NULL DEFAULT 0,
	size INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS versions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	path TEXT NOT NULL,
	name TEXT NOT NULL,
	file_path TEXT NOT NULL,
	size INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS audit (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	action TEXT NOT NULL,
	detail TEXT NOT NULL DEFAULT '',
	created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_trash_orig ON trash(user_id, original_path);
CREATE INDEX IF NOT EXISTS idx_versions_path ON versions(user_id, path);
CREATE INDEX IF NOT EXISTS idx_audit_time ON audit(created_at);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}
	// Migration: store app-password plaintext so the UI can show/copy it on demand.
	hasSecret, err := s.tableHasColumn("app_passwords", "secret")
	if err != nil {
		return err
	}
	if !hasSecret {
		if _, err := s.db.Exec(`ALTER TABLE app_passwords ADD COLUMN secret TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	return nil
}

// tableHasColumn reports whether table has the given column.
func (s *Store) tableHasColumn(table, col string) (bool, error) {
	rows, err := s.db.Query("SELECT 1 FROM pragma_table_info('" + table + "') WHERE name = '" + col + "'")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// DataDir returns the WebDAV data base directory.
func (s *Store) DataDir() string { return s.dataDir }

// userDir returns the private directory for a user.
func (s *Store) userDir(userID int64) string {
	return filepath.Join(s.dataDir, "users", fmt.Sprint(userID))
}

// UserRoot returns (creating if needed) the workspace root for a user.
func (s *Store) UserRoot(userID int64) string {
	d := filepath.Join(s.userDir(userID), "root")
	os.MkdirAll(d, 0o755)
	return d
}

// UserTrashDir returns (creating if needed) the trash directory for a user.
func (s *Store) UserTrashDir(userID int64) string {
	d := filepath.Join(s.userDir(userID), "trash")
	os.MkdirAll(d, 0o755)
	return d
}

// UserVersionDir returns (creating if needed) the versions directory for a user.
func (s *Store) UserVersionDir(userID int64) string {
	d := filepath.Join(s.userDir(userID), "versions")
	os.MkdirAll(d, 0o755)
	return d
}

// Resolve converts a WebDAV-relative path (starts with "/") for a user into a
// safe absolute filesystem path inside that user's workspace root.
func (s *Store) Resolve(userID int64, p string) (string, error) {
	root := s.UserRoot(userID)
	p = filepath.Clean("/" + strings.TrimPrefix(p, "/"))
	abs := filepath.Join(root, filepath.FromSlash(p))
	if abs != root && !strings.HasPrefix(abs, root+string(filepath.Separator)) {
		return "", fmt.Errorf("store: path escapes workspace: %s", p)
	}
	return abs, nil
}

func randToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// NewToken returns a random hex token.
func NewToken() string { return randToken(24) }

// ---- settings ----

// GetSetting returns a setting value or the provided default.
func (s *Store) GetSetting(key, def string) (string, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return def, nil
	}
	if err != nil {
		return "", err
	}
	return v, nil
}

// SetSetting upserts a setting.
func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(`INSERT INTO settings(key, value) VALUES(?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

// ---- audit ----

// AuditLog appends an audit entry.
func (s *Store) AuditLog(userID int64, action, detail string) error {
	_, err := s.db.Exec(`INSERT INTO audit(user_id, action, detail, created_at) VALUES(?,?,?,?)`,
		userID, action, detail, time.Now().UTC())
	return err
}

// AuditList returns audit entries, newest first.
func (s *Store) AuditList(limit int) ([]Audit, error) {
	rows, err := s.db.Query(`SELECT id, user_id, action, detail, created_at FROM audit ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Audit
	for rows.Next() {
		var a Audit
		if err := rows.Scan(&a.ID, &a.UserID, &a.Action, &a.Detail, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AuditDelete removes a single audit entry.
func (s *Store) AuditDelete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM audit WHERE id = ?`, id)
	return err
}

// AuditClear removes all audit entries.
func (s *Store) AuditClear() error {
	_, err := s.db.Exec(`DELETE FROM audit`)
	return err
}