package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"chinasclmdav/internal/fsys"
	"chinasclmdav/internal/store"
)

// handleListFiles lists a directory (or searches when ?q= is present).
func (s *Server) handleListFiles(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	q := r.URL.Query().Get("q")
	dir := r.URL.Query().Get("dir")
	abs, err := s.Store.Resolve(u.ID, dir)
	if err != nil {
		s.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		s.fail(w, http.StatusNotFound, "目录不存在")
		return
	}
	ignore := s.ignoreList()
	root := s.Store.UserRoot(u.ID)
	if q != "" {
		entries, err := fsys.WalkSearch(abs, q, 6)
		if err != nil {
			s.fail(w, http.StatusBadRequest, "搜索失败")
			return
		}
		filtered := entries[:0]
		for _, e := range entries {
			if fsys.IsHidden(filepath.Base(e.Path), ignore) {
				continue
			}
			e.Path = relURLPath(root, e.Path)
			filtered = append(filtered, e)
		}
		sort.Slice(filtered, func(i, j int) bool {
			if filtered[i].IsDir != filtered[j].IsDir {
				return filtered[i].IsDir
			}
			return filtered[i].Name < filtered[j].Name
		})
		s.ok(w, map[string]interface{}{
			"path":    dir,
			"search":  q,
			"entries": filtered,
			"crumbs":  fsys.Breadcrumbs(abs, root),
		})
		return
	}

	entries, err := fsys.List(abs, "")
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "读取目录失败")
		return
	}
	filtered := entries[:0]
	for _, e := range entries {
		if fsys.IsHidden(e.Name, ignore) {
			continue
		}
		e.Path = relURLPath(root, e.Path)
		filtered = append(filtered, e)
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].IsDir != filtered[j].IsDir {
			return filtered[i].IsDir
		}
		return filtered[i].Name < filtered[j].Name
	})
	s.ok(w, map[string]interface{}{
		"path":    dir,
		"entries": filtered,
		"crumbs":  fsys.Breadcrumbs(abs, root),
	})
}

func relURLPath(root, abs string) string {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "/"
	}
	return "/" + filepath.ToSlash(rel)
}

func (s *Server) ignoreList() []string {
	ig, _ := s.Store.GetSetting("ignore", ".git,.DS_Store,*.log")
	var out []string
	for _, p := range strings.Split(ig, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (s *Server) handleMkdir(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		Dir  string `json:"dir"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	if body.Name == "" {
		s.fail(w, http.StatusBadRequest, "名称不能为空")
		return
	}
	abs, err := s.Store.Resolve(u.ID, path.Join(body.Dir, body.Name))
	if err != nil {
		s.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := os.Mkdir(abs, 0o755); err != nil {
		s.fail(w, http.StatusBadRequest, "创建失败（目录可能已存在）")
		return
	}
	s.Store.AuditLog(u.ID, "mkdir", body.Dir+"/"+body.Name)
	s.ok(w, map[string]bool{"ok": true})
}

// handleDelete moves a file/folder to trash.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	abs, err := s.Store.Resolve(u.ID, body.Path)
	if err != nil {
		s.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := trash(u, s.Store, body.Path, abs); err != nil {
		s.fail(w, http.StatusInternalServerError, "删除失败")
		return
	}
	s.Store.AuditLog(u.ID, "delete", body.Path)
	s.ok(w, map[string]bool{"ok": true})
}

func trash(u *store.User, st *store.Store, rel, abs string) error {
	info, err := os.Stat(abs)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	dst := filepath.Join(st.UserTrashDir(u.ID), fmt.Sprintf("%d_%s", time.Now().UnixNano(), info.Name()))
	if err := os.Rename(abs, dst); err != nil {
		return err
	}
	size, _ := fsys.DirSize(dst, info.IsDir())
	_, err = st.AddTrash(u.ID, "/"+strings.TrimPrefix(rel, "/"), info.Name(), dst, info.IsDir(), size)
	if err != nil {
		_ = os.Rename(dst, abs)
		return err
	}
	return nil
}

func (s *Server) handleRename(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		Path string `json:"path"`
		New  string `json:"new"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	oldAbs, err := s.Store.Resolve(u.ID, body.Path)
	if err != nil {
		s.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	newAbs, err := s.Store.Resolve(u.ID, body.New)
	if err != nil {
		s.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := os.Rename(oldAbs, newAbs); err != nil {
		s.fail(w, http.StatusBadRequest, "重命名失败")
		return
	}
	s.Store.AuditLog(u.ID, "rename", body.Path+" -> "+body.New)
	s.ok(w, map[string]bool{"ok": true})
}

// handleDownload streams a file or zips a directory.
func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	p := r.URL.Query().Get("path")
	abs, err := s.Store.Resolve(u.ID, p)
	if err != nil {
		s.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	info, err := os.Stat(abs)
	if err != nil {
		s.fail(w, http.StatusNotFound, "文件不存在")
		return
	}
	if info.IsDir() {
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+url.PathEscape(info.Name())+".zip\"")
		fsys.Zip(w, abs, info.Name())
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+url.PathEscape(info.Name())+"\"")
	http.ServeFile(w, r, abs)
}

// handlePreview serves a file inline so the browser can render it (image,
// PDF, text, video, audio) instead of forcing a download.
func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	p := r.URL.Query().Get("path")
	abs, err := s.Store.Resolve(u.ID, p)
	if err != nil {
		s.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		s.fail(w, http.StatusNotFound, "文件不存在")
		return
	}
	ct := mime.TypeByExtension(filepath.Ext(info.Name()))
	if ct == "" {
		ct = "application/octet-stream"
	}
	if strings.HasPrefix(ct, "text/") {
		ct += "; charset=utf-8"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", "inline; filename=\""+url.PathEscape(info.Name())+"\"")
	http.ServeFile(w, r, abs)
}

// handleUpload stores an uploaded file.
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	if r.Method != "POST" {
		s.fail(w, http.StatusMethodNotAllowed, "")
		return
	}
	dir := r.URL.Query().Get("dir")
	absDir, err := s.Store.Resolve(u.ID, dir)
	if err != nil {
		s.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		s.fail(w, http.StatusInternalServerError, "目录不可用")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		s.fail(w, http.StatusBadRequest, "未找到文件")
		return
	}
	defer file.Close()
	dest := filepath.Join(absDir, filepath.Base(header.Filename))
	if err := os.WriteFile(dest, nil, 0o644); err != nil {
		s.fail(w, http.StatusInternalServerError, "写入失败")
		return
	}
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "写入失败")
		return
	}
	// snapshot old version if overwriting
	if fi, err := os.Stat(dest); err == nil && fi.Size() > 0 {
		_ = s.snapshotFile(u, path.Join(dir, header.Filename), dest, fi)
	}
	_, err = io.Copy(out, file)
	out.Close()
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "写入失败")
		return
	}
	s.Store.AuditLog(u.ID, "upload", path.Join(dir, header.Filename))
	s.ok(w, map[string]interface{}{"ok": true, "name": header.Filename})
}

func (s *Server) snapshotFile(u *store.User, rel, abs string, fi os.FileInfo) error {
	if !fi.Mode().IsRegular() {
		return nil
	}
	clean := strings.TrimPrefix(filepath.Clean("/"+strings.TrimPrefix(rel, "/")), "/")
	dir := filepath.Join(s.Store.UserVersionDir(u.ID), sanitize(clean))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	snap := filepath.Join(dir, time.Now().UTC().Format("20060102_150405.000000000")+"_"+fi.Name())
	if err := copyFile(abs, snap); err != nil {
		return err
	}
	_, err := s.Store.AddVersion(u.ID, clean, fi.Name(), snap, fi.Size())
	return err
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteString("_")
		}
	}
	return b.String()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}