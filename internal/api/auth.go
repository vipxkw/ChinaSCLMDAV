package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"chinasclmdav/internal/auth"
)

// LoginRequest is the login JSON body.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TotpCode string `json:"totp_code,omitempty"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.fail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	user, err := s.Store.GetUserByUsernameOrEmail(req.Username)
	if err != nil {
		s.fail(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	if !s.Store.CheckPassword(user, req.Password) {
		s.fail(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	if user.TotpEnabled {
		if req.TotpCode == "" {
			s.fail(w, http.StatusUnauthorized, "需要两步验证码")
			return
		}
		if !auth.VerifyTOTP(user.TotpSecret, req.TotpCode, time.Now()) {
			s.fail(w, http.StatusUnauthorized, "验证码错误")
			return
		}
	}
	sess, err := s.Store.CreateSession(user.ID, s.SessionTime)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "创建会话失败")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sess.Token,
		Path:     "/",
		MaxAge:   int(s.SessionTime.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	s.Store.AuditLog(user.ID, "login", "网页登录")
	s.ok(w, map[string]interface{}{
		"id":           user.ID,
		"username":     user.Username,
		"display_name": user.DisplayName,
		"email":        user.Email,
		"totp_enabled": user.TotpEnabled,
		"totp_forced":  user.TotpForced,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.fail(w, http.StatusMethodNotAllowed, "")
		return
	}
	c, err := r.Cookie("session")
	if err == nil {
		s.Store.DeleteSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true,
	})
	s.ok(w, map[string]bool{"ok": true})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	s.ok(w, map[string]interface{}{
		"id":           u.ID,
		"username":     u.Username,
		"display_name": u.DisplayName,
		"email":        u.Email,
		"totp_enabled": u.TotpEnabled,
		"totp_forced":  u.TotpForced,
	})
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	if err := s.Store.UpdateUser(u.ID, body.DisplayName, body.Email); err != nil {
		s.fail(w, http.StatusInternalServerError, "更新失败")
		return
	}
	s.Store.AuditLog(u.ID, "profile_update", "更新个人资料")
	s.ok(w, map[string]bool{"ok": true})
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	if !s.Store.CheckPassword(u, body.OldPassword) {
		s.fail(w, http.StatusForbidden, "原密码错误")
		return
	}
	if len(body.NewPassword) < 4 {
		s.fail(w, http.StatusBadRequest, "密码太短")
		return
	}
	if err := s.Store.UpdatePassword(u.ID, body.NewPassword); err != nil {
		s.fail(w, http.StatusInternalServerError, "修改失败")
		return
	}
	s.Store.AuditLog(u.ID, "password_change", "修改密码")
	s.ok(w, map[string]bool{"ok": true})
}

// TOTP endpoints

func (s *Server) handleTotpEnable(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	if u.TotpSecret == "" {
		s.fail(w, http.StatusBadRequest, "请先获取密钥")
		return
	}
	if !auth.VerifyTOTP(u.TotpSecret, body.Code, time.Now()) {
		s.fail(w, http.StatusBadRequest, "验证码错误")
		return
	}
	if err := s.Store.SetTotpEnabled(u.ID, true); err != nil {
		s.fail(w, http.StatusInternalServerError, "启用失败")
		return
	}
	s.Store.AuditLog(u.ID, "totp_enable", "启用两步验证")
	s.ok(w, map[string]bool{"ok": true})
}

func (s *Server) handleTotpDisable(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	if !auth.VerifyTOTP(u.TotpSecret, body.Code, time.Now()) {
		s.fail(w, http.StatusBadRequest, "验证码错误")
		return
	}
	if err := s.Store.SetTotpEnabled(u.ID, false); err != nil {
		s.fail(w, http.StatusInternalServerError, "禁用失败")
		return
	}
	s.Store.AuditLog(u.ID, "totp_disable", "禁用两步验证")
	s.ok(w, map[string]bool{"ok": true})
}

func (s *Server) handleTotpSecret(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	secret := auth.GenerateTOTPSecret()
	if err := s.Store.UpdateTotpSecret(u.ID, secret); err != nil {
		s.fail(w, http.StatusInternalServerError, "生成密钥失败")
		return
	}
	uri := auth.OTPURI(u.Username, "ChinaSCLM DAV", secret)
	s.ok(w, map[string]string{"secret": secret, "uri": uri})
}

func (s *Server) handleTotpForce(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		Forced bool `json:"forced"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	if err := s.Store.SetTotpForced(u.ID, body.Forced); err != nil {
		s.fail(w, http.StatusInternalServerError, "设置失败")
		return
	}
	s.Store.AuditLog(u.ID, "totp_policy", "设置TOTP强制策略")
	s.ok(w, map[string]bool{"ok": true})
}

// Settings

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	_ = u
	ignore, _ := s.Store.GetSetting("ignore", ".git,.DS_Store,*.log")
	lang, _ := s.Store.GetSetting("lang", "zh")
	compress, _ := s.Store.GetSetting("compress", "low")
	s.ok(w, map[string]string{
		"ignore":   ignore,
		"lang":     lang,
		"compress": compress,
	})
}

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	var body struct {
		Ignore   string `json:"ignore,omitempty"`
		Lang     string `json:"lang,omitempty"`
		Compress string `json:"compress,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.fail(w, http.StatusBadRequest, "无效请求")
		return
	}
	if body.Ignore != "" {
		s.Store.SetSetting("ignore", body.Ignore)
	}
	if body.Lang != "" {
		s.Store.SetSetting("lang", body.Lang)
	}
	if body.Compress != "" {
		s.Store.SetSetting("compress", body.Compress)
	}
	s.Store.AuditLog(u.ID, "settings", "更新设置")
	s.ok(w, map[string]bool{"ok": true})
}

// Audit

func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	u, ok := s.authUser(r)
	if !ok {
		s.fail(w, http.StatusUnauthorized, "未登录")
		return
	}
	_ = u
	if r.Method == http.MethodDelete {
		idStr := r.URL.Query().Get("id")
		if idStr != "" {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				s.fail(w, http.StatusBadRequest, "无效的日志 ID")
				return
			}
			if err := s.Store.AuditDelete(id); err != nil {
				s.fail(w, http.StatusInternalServerError, "删除失败")
				return
			}
		} else {
			if err := s.Store.AuditClear(); err != nil {
				s.fail(w, http.StatusInternalServerError, "清空失败")
				return
			}
		}
		s.ok(w, map[string]bool{"ok": true})
		return
	}
	entries, err := s.Store.AuditList(1000)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "查询失败")
		return
	}
	s.ok(w, entries)
}