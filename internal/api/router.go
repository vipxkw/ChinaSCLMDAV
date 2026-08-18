package api

import (
	"net/http"
	"strings"

	"chinasclmdav/internal/dav"
	"chinasclmdav/internal/static"
	"chinasclmdav/internal/store"
)

// NewRouter builds the full HTTP mux.
func NewRouter(st *store.Store) http.Handler {
	mux := http.NewServeMux()
	s := New(st)

	// Public
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/logout", s.handleLogout)

	// Authenticated
	mux.HandleFunc("/api/me", s.requireAuth(s.handleMe))
	mux.HandleFunc("/api/profile", s.requireAuth(s.handleUpdateProfile))
	mux.HandleFunc("/api/password", s.requireAuth(s.handleChangePassword))
	mux.HandleFunc("/api/dashboard", s.requireAuth(s.handleDashboard))
	mux.HandleFunc("/api/files", s.requireAuth(s.handleListFiles))
	mux.HandleFunc("/api/mkdir", s.requireAuth(s.handleMkdir))
	mux.HandleFunc("/api/delete", s.requireAuth(s.handleDelete))
	mux.HandleFunc("/api/rename", s.requireAuth(s.handleRename))
	mux.HandleFunc("/api/download", s.requireAuth(s.handleDownload))
	mux.HandleFunc("/api/preview", s.requireAuth(s.handlePreview))
	mux.HandleFunc("/api/upload", s.requireAuth(s.handleUpload))
	mux.HandleFunc("/api/trash", s.requireAuth(handleTrashRouter(s)))
	mux.HandleFunc("/api/trash/", s.requireAuth(handleTrashRouter(s)))
	mux.HandleFunc("/api/versions", s.requireAuth(handleVersionRouter(s)))
	mux.HandleFunc("/api/versions/", s.requireAuth(handleVersionRouter(s)))
	mux.HandleFunc("/api/shares", s.requireAuth(handleShareRouter(s)))
	mux.HandleFunc("/api/app-passwords", s.requireAuth(handleAppPwdRouter(s)))
	mux.HandleFunc("/api/settings", s.requireAuth(handleSettingsRouter(s)))
	mux.HandleFunc("/api/audit", s.requireAuth(s.handleAudit))

	// TOTP
	mux.HandleFunc("/api/totp/secret", s.requireAuth(s.handleTotpSecret))
	mux.HandleFunc("/api/totp/enable", s.requireAuth(s.handleTotpEnable))
	mux.HandleFunc("/api/totp/disable", s.requireAuth(s.handleTotpDisable))
	mux.HandleFunc("/api/totp/force", s.requireAuth(s.handleTotpForce))
	mux.HandleFunc("/api/refresh", s.requireAuth(s.handleRefresh))

	// Public share
	mux.HandleFunc("/s/", s.handlePublicShare)

	// WebDAV
	mux.Handle("/dav/", dav.New(st))

	// SPA static files — catch-all
	mux.HandleFunc("/", spaHandler)

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// handleTrashRouter dispatches sub-routes for /api/trash/...
func handleTrashRouter(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/api/trash")
		switch p {
		case "/restore":
			s.handleRestoreTrash(w, r)
		case "/purge":
			s.handlePurgeTrash(w, r)
		case "/empty":
			s.handleEmptyTrash(w, r)
		default:
			s.handleListTrash(w, r)
		}
	}
}

func handleVersionRouter(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/api/versions")
		switch p {
		case "/download":
			s.handleDownloadVersion(w, r)
		case "/delete":
			s.handleDeleteVersion(w, r)
		default:
			s.handleListVersions(w, r)
		}
	}
}

func handleShareRouter(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			s.handleCreateShare(w, r)
		case "DELETE":
			s.handleDeleteShare(w, r)
		default:
			s.handleListShares(w, r)
		}
	}
}

func handleAppPwdRouter(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			s.handleCreateAppPassword(w, r)
		case "DELETE":
			s.handleDeleteAppPassword(w, r)
		default:
			s.handleListAppPasswords(w, r)
		}
	}
}

func handleSettingsRouter(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			s.handleUpdateSettings(w, r)
		default:
			s.handleGetSettings(w, r)
		}
	}
}

// spaHandler serves the single-page app or falls through to 404.
func spaHandler(w http.ResponseWriter, r *http.Request) {
	// If the path is an API path, return 404 (shouldn't reach here).
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/dav/") || strings.HasPrefix(r.URL.Path, "/s/") {
		http.NotFound(w, r)
		return
	}
	// Serve embedded static files (SPA).
	static.Handler().ServeHTTP(w, r)
}