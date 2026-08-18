package store

import (
	"database/sql"
	"time"
)

// CreateShare creates a public share for a path.
func (s *Store) CreateShare(userID int64, path, name string, size int64, expiresAt *time.Time) (*Share, error) {
	sh := &Share{
		UserID:    userID,
		Token:     NewToken(),
		Path:      path,
		Name:      name,
		Size:      size,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}
	_, err := s.db.Exec(`INSERT INTO shares(user_id,token,path,name,size,expires_at,created_at) VALUES(?,?,?,?,?,?,?)`,
		sh.UserID, sh.Token, sh.Path, sh.Name, sh.Size, sh.ExpiresAt, sh.CreatedAt)
	if err != nil {
		return nil, err
	}
	return sh, nil
}

// GetShareByToken returns a share by token (filters expired).
func (s *Store) GetShareByToken(token string) (*Share, error) {
	sh := &Share{}
	err := s.db.QueryRow(`SELECT id,user_id,token,path,name,size,expires_at,download_count,created_at
		FROM shares WHERE token = ?`, token).Scan(
		&sh.ID, &sh.UserID, &sh.Token, &sh.Path, &sh.Name, &sh.Size, &sh.ExpiresAt,
		&sh.DownloadCount, &sh.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if sh.ExpiresAt != nil && time.Now().After(*sh.ExpiresAt) {
		return nil, ErrNotFound
	}
	return sh, nil
}

// ListShares returns all shares for a user.
func (s *Store) ListShares(userID int64) ([]Share, error) {
	rows, err := s.db.Query(`SELECT id,user_id,token,path,name,size,expires_at,download_count,created_at
		FROM shares WHERE user_id = ? ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Share
	for rows.Next() {
		var sh Share
		if err := rows.Scan(&sh.ID, &sh.UserID, &sh.Token, &sh.Path, &sh.Name, &sh.Size,
			&sh.ExpiresAt, &sh.DownloadCount, &sh.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sh)
	}
	return out, rows.Err()
}

// DeleteShare removes a share.
func (s *Store) DeleteShare(id, userID int64) error {
	res, err := s.db.Exec(`DELETE FROM shares WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// IncrementShareDownload increments the download counter.
func (s *Store) IncrementShareDownload(id int64) error {
	_, err := s.db.Exec(`UPDATE shares SET download_count = download_count + 1 WHERE id = ?`, id)
	return err
}

// ---- trash ----

// AddTrash records a deleted item and returns its id.
func (s *Store) AddTrash(userID int64, originalPath, name, trashPath string, isDir bool, size int64) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO trash(user_id,original_path,name,trash_path,is_dir,size,deleted_at)
		VALUES(?,?,?,?,?,?,?)`, userID, originalPath, name, trashPath, isDir, size, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListTrash returns trash items for a user.
func (s *Store) ListTrash(userID int64) ([]TrashItem, error) {
	rows, err := s.db.Query(`SELECT id,original_path,name,trash_path,is_dir,size,deleted_at
		FROM trash WHERE user_id = ? ORDER BY deleted_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TrashItem
	for rows.Next() {
		var t TrashItem
		if err := rows.Scan(&t.ID, &t.OriginalPath, &t.Name, &t.TrashPath, &t.IsDir, &t.Size, &t.DeletedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetTrash returns a trash item by id.
func (s *Store) GetTrash(id, userID int64) (*TrashItem, error) {
	t := &TrashItem{}
	err := s.db.QueryRow(`SELECT id,original_path,name,trash_path,is_dir,size,deleted_at
		FROM trash WHERE id = ? AND user_id = ?`, id, userID).Scan(
		&t.ID, &t.OriginalPath, &t.Name, &t.TrashPath, &t.IsDir, &t.Size, &t.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return t, err
}

// DeleteTrash removes a trash record.
func (s *Store) DeleteTrash(id, userID int64) error {
	res, err := s.db.Exec(`DELETE FROM trash WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ---- versions ----

// AddVersion records a file version.
func (s *Store) AddVersion(userID int64, path, name, filePath string, size int64) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO versions(user_id,path,name,file_path,size,created_at)
		VALUES(?,?,?,?,?,?)`, userID, path, name, filePath, size, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListVersions returns all versions for a path.
func (s *Store) ListVersions(userID int64, path string) ([]Version, error) {
	rows, err := s.db.Query(`SELECT id,path,name,file_path,size,created_at
		FROM versions WHERE user_id = ? AND path = ? ORDER BY created_at DESC`, userID, path)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Version
	for rows.Next() {
		var v Version
		if err := rows.Scan(&v.ID, &v.Path, &v.Name, &v.FilePath, &v.Size, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// GetVersion returns a version by id.
func (s *Store) GetVersion(id, userID int64) (*Version, error) {
	v := &Version{}
	err := s.db.QueryRow(`SELECT id,path,name,file_path,size,created_at
		FROM versions WHERE id = ? AND user_id = ?`, id, userID).Scan(
		&v.ID, &v.Path, &v.Name, &v.FilePath, &v.Size, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return v, err
}

// DeleteVersion removes a version record.
func (s *Store) DeleteVersion(id, userID int64) error {
	res, err := s.db.Exec(`DELETE FROM versions WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// PurgeVersionsForPath removes all versions for a path.
func (s *Store) PurgeVersionsForPath(userID int64, path string) error {
	_, err := s.db.Exec(`DELETE FROM versions WHERE user_id = ? AND path = ?`, userID, path)
	return err
}