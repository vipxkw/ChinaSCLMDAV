package api

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"chinasclmdav/internal/fsys"
)

// handleDashboard returns storage stats, categories and recent files.
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	root := s.Store.UserRoot(u.ID)
	categories, totalSize, fileCount, err := fsys.Categories(root)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "统计失败")
		return
	}
	dirCount := int64(0)
	ignore := s.ignoreList()
	filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() {
			return nil
		}
		if fsys.IsHidden(fi.Name(), ignore) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
		}
		dirCount++
		return nil
	})

	// recent files: walk up to depth 4 collecting files sorted by mtime.
	var recent []fsys.Entry
	walkRecent(root, root, 4, ignore, &recent)
	sort.Slice(recent, func(i, j int) bool { return recent[i].ModTime > recent[j].ModTime })
	if len(recent) > 20 {
		recent = recent[:20]
	}
	for i := range recent {
		recent[i].Path = relURLPath(root, recent[i].Path)
	}
	s.ok(w, map[string]interface{}{
		"total_size":    totalSize,
		"file_count":    fileCount,
		"dir_count":     dirCount,
		"categories":    categories,
		"recent":        recent,
	})
}

func walkRecent(root, dir string, depth int, ignore []string, out *[]fsys.Entry) {
	if depth <= 0 {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, de := range entries {
		name := de.Name()
		if fsys.IsHidden(name, ignore) {
			continue
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		if de.IsDir() {
			walkRecent(root, filepath.Join(dir, name), depth-1, ignore, out)
			continue
		}
		*out = append(*out, fsys.Entry{
			Name:    name,
			Path:    filepath.Join(dir, name),
			IsDir:   false,
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
		})
	}
}

// handleHealth is a lightweight health check.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.ok(w, map[string]string{"status": "ok", "time": time.Now().Format(time.RFC3339)})
}

// handleRefresh is a no-op used by the UI to force a reload signal.
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	s.ok(w, map[string]bool{"ok": true})
}