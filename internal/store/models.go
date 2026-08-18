package store

import "time"

// User is an account that can log in to the web UI and WebDAV.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"`
	TotpSecret   string    `json:"-"`
	TotpEnabled  bool      `json:"totp_enabled"`
	TotpForced   bool      `json:"totp_forced"`
	CreatedAt    time.Time `json:"created_at"`
}

// Session is a login session bound to a user.
type Session struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AppPassword is an application-specific password used for WebDAV clients.
type AppPassword struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"-"`
	AppName   string    `json:"app_name"`
	Secret    string    `json:"password"`
	Hash      string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// Share is a public, expiring link to a file.
type Share struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"-"`
	Token        string     `json:"token"`
	Path         string     `json:"path"`
	Name         string     `json:"name"`
	Size         int64      `json:"size"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	DownloadCount int64     `json:"download_count"`
	CreatedAt    time.Time  `json:"created_at"`
}

// TrashItem is a deleted file/folder awaiting restore or purge.
type TrashItem struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"-"`
	OriginalPath string    `json:"original_path"`
	Name         string    `json:"name"`
	TrashPath    string    `json:"-"`
	IsDir        bool      `json:"is_dir"`
	Size         int64     `json:"size"`
	DeletedAt    time.Time `json:"deleted_at"`
}

// Version is a snapshot of a file taken before it was overwritten.
type Version struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"-"`
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	FilePath   string    `json:"-"`
	Size       int64     `json:"size"`
	CreatedAt  time.Time `json:"created_at"`
}

// Audit is an audit-log entry.
type Audit struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}