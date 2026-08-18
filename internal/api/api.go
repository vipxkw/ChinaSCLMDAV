package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"chinasclmdav/internal/store"
)

// ctxKey is the request context key holding the authenticated user.
type ctxKey int

const userKey ctxKey = 0

// Server holds the HTTP API handlers.
type Server struct {
	Store       *store.Store
	SessionTime time.Duration
	PublicURL   string
}

// New creates an API server.
func New(st *store.Store) *Server {
	return &Server{Store: st, SessionTime: 24 * time.Hour}
}

func (s *Server) fail(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *Server) ok(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (s *Server) authUser(r *http.Request) (*store.User, bool) {
	u, _ := r.Context().Value(userKey).(*store.User)
	if u == nil {
		return nil, false
	}
	return u, true
}

func withUser(ctx context.Context, u *store.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// requireAuth wraps a handler, requiring a valid session cookie.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := s.authViaCookie(r)
		if !ok {
			s.fail(w, http.StatusUnauthorized, "未登录")
			return
		}
		next(w, r.WithContext(withUser(r.Context(), u)))
	}
}

func (s *Server) authViaCookie(r *http.Request) (*store.User, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		return nil, false
	}
	sess, err := s.Store.GetSession(c.Value)
	if err != nil {
		return nil, false
	}
	u, err := s.Store.GetUser(sess.UserID)
	if err != nil {
		return nil, false
	}
	return u, true
}

// optionalAuth tries to authenticate but does not reject unauthenticated
// requests.
func (s *Server) optionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if u, ok := s.authViaCookie(r); ok {
			r = r.WithContext(withUser(r.Context(), u))
		}
		next(w, r)
	}
}