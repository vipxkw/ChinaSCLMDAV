package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chinasclmdav/internal/api"
	"chinasclmdav/internal/store"
)

// envOr returns the value of environment variable key, or def if unset.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	var (
		dataDir   = flag.String("data", envOr("CHINASCLMDAV_DATA", "./data"), "data directory (stores user files + sqlite db)")
		dbPath    = flag.String("db", envOr("CHINASCLMDAV_DB", ""), "sqlite database path (default: <data>/chinasclmdav.db)")
		addr      = flag.String("listen", envOr("CHINASCLMDAV_LISTEN", ":8080"), "http listen address")
		publicURL = flag.String("public-url", envOr("CHINASCLMDAV_PUBLIC_URL", "http://localhost:8080"), "public base url used in WebDAV hints")
		seedUser  = flag.String("seed-user", envOr("CHINASCLMDAV_SEED_USER", "admin"), "username to seed on first run")
		seedEmail = flag.String("seed-email", envOr("CHINASCLMDAV_SEED_EMAIL", "admin@qq.com"), "email for the seeded user")
		seedName  = flag.String("seed-display", envOr("CHINASCLMDAV_SEED_NAME", "admin"), "display name for the seeded user")
		seedPass  = flag.String("seed-pass", envOr("CHINASCLMDAV_SEED_PASS", "admin123"), "password for the seeded user (only on first run)")
	)
	flag.Parse()

	if *dbPath == "" {
		*dbPath = filepath.Join(*dataDir, "chinasclmdav.db")
	}

	st, err := store.Open(store.Config{DataDir: *dataDir, DBPath: *dbPath})
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	// Seed the first admin user on a fresh database.
	if err := ensureSeedUser(st, *seedUser, *seedEmail, *seedName, *seedPass); err != nil {
		log.Fatalf("seed user: %v", err)
	}

	router := api.NewRouter(st)

	srv := &http.Server{
		Addr:              *addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("ChinaSCLM DAV WebDAV server listening on %s", *addr)
	log.Printf("WebDAV endpoint: %s/dav/  (use an app password)", strings.TrimRight(*publicURL, "/"))
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// ensureSeedUser creates an initial user only when the users table is empty.
func ensureSeedUser(st *store.Store, username, email, displayName, password string) error {
	existing, err := st.GetUserByUsernameOrEmail(username)
	if err == nil && existing != nil {
		return nil
	}
	// No user exists yet → create the seed account.
	if password == "" {
		return fmt.Errorf("first run requires a seed password (use --seed-pass), or existing data at %s", st.DataDir())
	}
	_, err = st.CreateUser(username, email, displayName, password)
	if err != nil {
		return err
	}
	log.Printf("seeded user %q (%s)", username, email)
	if os.Getenv("CHINASCLMDAV_SEED_FORCE_REMIND") != "" {
		// no-op hook kept for environments that want to log a hint
	}
	return nil
}