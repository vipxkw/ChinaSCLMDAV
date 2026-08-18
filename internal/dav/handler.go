package dav

import (
	"net/http"

	webdav "golang.org/x/net/webdav"

	"chinasclmdav/internal/store"
)

// Handler serves the WebDAV protocol under /dav using Basic auth.
// Passwords are accepted as either the user's login password (when TOTP is
// not enforced) or any valid application password (always).
type Handler struct {
	store *store.Store
}

// New returns an http.Handler for the WebDAV endpoint.
func New(st *store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := basicAuthenticate(st, r)
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="ChinaSCLM DAV"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h := &webdav.Handler{
			Prefix:     "/dav",
			FileSystem: NewFileSystem(st, user),
			LockSystem: webdav.NewMemLS(),
		}
		h.ServeHTTP(w, r)
	})
}

// BasicAuthenticate validates Basic auth credentials.
func basicAuthenticate(st *store.Store, r *http.Request) (*store.User, bool) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, false
	}
	user, err := st.GetUserByUsernameOrEmail(username)
	if err != nil {
		return nil, false
	}
	// App passwords always work (they bypass TOTP).
	if ok, _ := st.VerifyAppPassword(user.ID, password); ok {
		return user, true
	}
	// The login password only works when TOTP is not enabled.
	if user.TotpEnabled {
		return nil, false
	}
	if st.CheckPassword(user, password) {
		return user, true
	}
	return nil, false
}