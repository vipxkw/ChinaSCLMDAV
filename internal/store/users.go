package store

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user with a bcrypt-hashed password.
func (s *Store) CreateUser(username, email, displayName, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &User{
		Username:     username,
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	res, err := s.db.Exec(`INSERT INTO users(username,email,display_name,password_hash,created_at) VALUES(?,?,?,?,?)`,
		u.Username, u.Email, u.DisplayName, u.PasswordHash, u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.ID, _ = res.LastInsertId()
	return u, nil
}

// GetUserByUsernameOrEmail looks up a user by username or email.
func (s *Store) GetUserByUsernameOrEmail(login string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(`SELECT id,username,email,display_name,password_hash,totp_secret,totp_enabled,totp_forced,created_at
		FROM users WHERE username = ? OR email = ?`, login, login).Scan(
		&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.PasswordHash,
		&u.TotpSecret, &u.TotpEnabled, &u.TotpForced, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return u, err
}

// GetUser returns a user by ID.
func (s *Store) GetUser(id int64) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(`SELECT id,username,email,display_name,password_hash,totp_secret,totp_enabled,totp_forced,created_at
		FROM users WHERE id = ?`, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.PasswordHash,
		&u.TotpSecret, &u.TotpEnabled, &u.TotpForced, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return u, err
}

// UpdateUser updates username, display name and email.
func (s *Store) UpdateUser(id int64, username, displayName, email string) error {
	_, err := s.db.Exec(`UPDATE users SET username = ?, display_name = ?, email = ? WHERE id = ?`, username, displayName, email, id)
	return err
}

// UpdatePassword updates the password hash.
func (s *Store) UpdatePassword(id int64, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, string(hash), id)
	return err
}

// CheckPassword verifies a password against the stored hash.
func (s *Store) CheckPassword(u *User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

// UpdateTotpSecret sets the TOTP secret and optionally enables TOTP.
func (s *Store) UpdateTotpSecret(id int64, secret string) error {
	_, err := s.db.Exec(`UPDATE users SET totp_secret = ? WHERE id = ?`, secret, id)
	return err
}

// SetTotpEnabled enables or disables TOTP for a user.
func (s *Store) SetTotpEnabled(id int64, enabled bool) error {
	_, err := s.db.Exec(`UPDATE users SET totp_enabled = ? WHERE id = ?`, enabled, id)
	return err
}

// SetTotpForced sets the TOTP forced policy for a user.
func (s *Store) SetTotpForced(id int64, forced bool) error {
	_, err := s.db.Exec(`UPDATE users SET totp_forced = ? WHERE id = ?`, forced, id)
	return err
}

// ---- sessions ----

// CreateSession creates a new session token for a user.
func (s *Store) CreateSession(userID int64, expiry time.Duration) (*Session, error) {
	tok := NewToken()
	sess := &Session{
		Token:     tok,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(expiry),
	}
	_, err := s.db.Exec(`INSERT INTO sessions(token,user_id,created_at,expires_at) VALUES(?,?,?,?)`,
		sess.Token, sess.UserID, sess.CreatedAt, sess.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// GetSession returns a session by token, ignoring expired ones.
func (s *Store) GetSession(token string) (*Session, error) {
	sess := &Session{}
	err := s.db.QueryRow(`SELECT token,user_id,created_at,expires_at FROM sessions
		WHERE token = ? AND expires_at > ?`, token, time.Now().UTC()).Scan(
		&sess.Token, &sess.UserID, &sess.CreatedAt, &sess.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return sess, err
}

// DeleteSession removes a session.
func (s *Store) DeleteSession(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

// CleanSessions removes all expired sessions.
func (s *Store) CleanSessions() error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE expires_at <= ?`, time.Now().UTC())
	return err
}

// ---- app passwords ----

// CreateAppPassword creates an app password and returns the plaintext.
func (s *Store) CreateAppPassword(userID int64, appName string) (string, error) {
	plain := randToken(16)
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	_, err = s.db.Exec(`INSERT INTO app_passwords(user_id,app_name,secret,hash,created_at) VALUES(?,?,?,?,?)`,
		userID, appName, plain, string(hash), time.Now().UTC())
	if err != nil {
		return "", err
	}
	return plain, nil
}

// ListAppPasswords returns all app passwords for a user (plaintext included for the UI).
func (s *Store) ListAppPasswords(userID int64) ([]AppPassword, error) {
	rows, err := s.db.Query(`SELECT id,app_name,secret,created_at FROM app_passwords WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AppPassword
	for rows.Next() {
		var ap AppPassword
		if err := rows.Scan(&ap.ID, &ap.AppName, &ap.Secret, &ap.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, ap)
	}
	return out, rows.Err()
}

// DeleteAppPassword removes an app password.
func (s *Store) DeleteAppPassword(id, userID int64) error {
	res, err := s.db.Exec(`DELETE FROM app_passwords WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// VerifyAppPassword checks a password against the stored app passwords.
func (s *Store) VerifyAppPassword(userID int64, password string) (bool, error) {
	rows, err := s.db.Query(`SELECT hash FROM app_passwords WHERE user_id = ?`, userID)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return false, err
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
			return true, nil
		}
	}
	return false, nil
}