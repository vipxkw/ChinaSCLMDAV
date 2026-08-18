package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// handleListShares lists all shares for the user.
func (s *Server) handleListShares(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	shares, err := s.Store.ListShares(u.ID)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "查询失败")
		return
	}
	s.ok(w, shares)
}

// handleCreateShare creates a share for a path.
func (s *Server) handleCreateShare(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		Path      string `json:"path"`
		ExpiresIn int64  `json:"expires_in"` // hours; 0 = never
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
	info, err := os.Stat(abs)
	if err != nil {
		s.fail(w, http.StatusNotFound, "文件不存在")
		return
	}
	var exp *time.Time
	if body.ExpiresIn > 0 {
		t := time.Now().UTC().Add(time.Duration(body.ExpiresIn) * time.Hour)
		exp = &t
	}
	sh, err := s.Store.CreateShare(u.ID, "/"+strings.TrimPrefix(body.Path, "/"), info.Name(), info.Size(), exp)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "创建失败")
		return
	}
	s.Store.AuditLog(u.ID, "share_create", body.Path)
	s.ok(w, sh)
}

// handleDeleteShare removes a share.
func (s *Server) handleDeleteShare(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	if err := s.Store.DeleteShare(body.ID, u.ID); err != nil {
		s.fail(w, http.StatusNotFound, "分享不存在")
		return
	}
	s.Store.AuditLog(u.ID, "share_delete", "")
	s.ok(w, map[string]bool{"ok": true})
}

// handlePublicShare serves a public share (no auth).
func (s *Server) handlePublicShare(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(path.Base(r.URL.Path), "")
	// Token is the last path segment.
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 2 {
		token = parts[len(parts)-1]
	}
	sh, err := s.Store.GetShareByToken(token)
	if err != nil {
		s.fail(w, http.StatusNotFound, "分享不存在或已过期")
		return
	}
	abs, err := s.Store.Resolve(sh.UserID, sh.Path)
	if err != nil || !fileExists(abs) {
		s.fail(w, http.StatusNotFound, "文件不存在")
		return
	}
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		s.fail(w, http.StatusNotFound, "仅支持分享文件")
		return
	}
	s.Store.IncrementShareDownload(sh.ID)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+info.Name()+"\"")
	http.ServeFile(w, r, abs)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// ---- trash ----

func (s *Server) handleListTrash(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	items, err := s.Store.ListTrash(u.ID)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "查询失败")
		return
	}
	s.ok(w, items)
}

// handleRestoreTrash moves an item back to its original path.
func (s *Server) handleRestoreTrash(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	item, err := s.Store.GetTrash(body.ID, u.ID)
	if err != nil {
		s.fail(w, http.StatusNotFound, "条目不存在")
		return
	}
	dst, err := s.Store.Resolve(u.ID, item.OriginalPath)
	if err != nil {
		s.fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := os.MkdirAll(path.Dir(dst), 0o755); err != nil {
		s.fail(w, http.StatusInternalServerError, "恢复失败")
		return
	}
	if fileExists(dst) {
		s.fail(w, http.StatusConflict, "目标位置已存在同名项")
		return
	}
	if err := os.Rename(item.TrashPath, dst); err != nil {
		s.fail(w, http.StatusInternalServerError, "恢复失败")
		return
	}
	s.Store.DeleteTrash(item.ID, u.ID)
	s.Store.AuditLog(u.ID, "trash_restore", item.OriginalPath)
	s.ok(w, map[string]bool{"ok": true})
}

// handlePurgeTrash permanently deletes an item.
func (s *Server) handlePurgeTrash(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	item, err := s.Store.GetTrash(body.ID, u.ID)
	if err != nil {
		s.fail(w, http.StatusNotFound, "条目不存在")
		return
	}
	if err := os.RemoveAll(item.TrashPath); err != nil {
		s.fail(w, http.StatusInternalServerError, "删除失败")
		return
	}
	s.Store.DeleteTrash(item.ID, u.ID)
	s.Store.AuditLog(u.ID, "trash_purge", item.OriginalPath)
	s.ok(w, map[string]bool{"ok": true})
}

// handleEmptyTrash permanently deletes everything in the trash.
func (s *Server) handleEmptyTrash(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	items, err := s.Store.ListTrash(u.ID)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "查询失败")
		return
	}
	for _, it := range items {
		os.RemoveAll(it.TrashPath)
		s.Store.DeleteTrash(it.ID, u.ID)
	}
	s.Store.AuditLog(u.ID, "trash_empty", "清空回收站")
	s.ok(w, map[string]bool{"ok": true})
}

// ---- versions ----

func (s *Server) handleListVersions(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	p := r.URL.Query().Get("path")
	versions, err := s.Store.ListVersions(u.ID, strings.TrimPrefix(p, "/"))
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "查询失败")
		return
	}
	s.ok(w, versions)
}

// handleDownloadVersion streams a version snapshot.
func (s *Server) handleDownloadVersion(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	idStr := r.URL.Query().Get("id")
	var id int64
	_, _ = fmt.Sscan(idStr, &id)
	v, err := s.Store.GetVersion(id, u.ID)
	if err != nil {
		s.fail(w, http.StatusNotFound, "版本不存在")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+v.Name+"\"")
	http.ServeFile(w, r, v.FilePath)
}

// handleDeleteVersion removes a version snapshot.
func (s *Server) handleDeleteVersion(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	v, err := s.Store.GetVersion(body.ID, u.ID)
	if err != nil {
		s.fail(w, http.StatusNotFound, "版本不存在")
		return
	}
	os.Remove(v.FilePath)
	s.Store.DeleteVersion(body.ID, u.ID)
	s.Store.AuditLog(u.ID, "version_delete", v.Path)
	s.ok(w, map[string]bool{"ok": true})
}

// ---- app passwords ----

func (s *Server) handleListAppPasswords(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	apps, err := s.Store.ListAppPasswords(u.ID)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "查询失败")
		return
	}
	s.ok(w, apps)
}

func (s *Server) handleCreateAppPassword(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		AppName string `json:"app_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	if body.AppName == "" {
		s.fail(w, http.StatusBadRequest, "应用名称不能为空")
		return
	}
	plain, err := s.Store.CreateAppPassword(u.ID, body.AppName)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "创建失败")
		return
	}
	s.Store.AuditLog(u.ID, "app_password_create", body.AppName)
	s.ok(w, map[string]string{"app_name": body.AppName, "password": plain})
}

func (s *Server) handleDeleteAppPassword(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	if err := s.Store.DeleteAppPassword(body.ID, u.ID); err != nil {
		s.fail(w, http.StatusNotFound, "应用密码不存在")
		return
	}
	s.Store.AuditLog(u.ID, "app_password_delete", "")
	s.ok(w, map[string]bool{"ok": true})
}