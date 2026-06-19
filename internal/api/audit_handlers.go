package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prejudice-studio/twilight/internal/store"
)

// audit 是写入操作审计日志的便捷方法。category 为 "admin" / "user" / "system"。
// AuditLog.enabled=false 时静默跳过记录。从 current(r) 提取操作者身份。
func (a *App) audit(r *http.Request, action, category string, targetUID int64, detail map[string]any) {
	p := current(r)
	a.auditEntry(r, p.User.UID, p.User.Username, action, category, targetUID, detail)
}

// auditWithUser 用于登录等尚无会话上下文但已知用户身份的路径。
// 避免因为 AuthPublic 接口中 current(r) 返回零值导致审计日志 uid=0 / username=""。
func (a *App) auditWithUser(r *http.Request, uid int64, username, action, category string, targetUID int64, detail map[string]any) {
	a.auditEntry(r, uid, username, action, category, targetUID, detail)
}

func (a *App) auditEntry(r *http.Request, uid int64, username, action, category string, targetUID int64, detail map[string]any) {
	a.auditEntryIP(a.clientIP(r), uid, username, action, category, targetUID, detail)
}

// auditEntryIP 是不依赖 *http.Request 的审计写入入口，供没有 HTTP 上下文的路径
// （如 Telegram Bot 命令）使用，IP 由调用方显式传入（如 "telegram"）。
func (a *App) auditEntryIP(ip string, uid int64, username, action, category string, targetUID int64, detail map[string]any) {
	cfg := a.cfg()
	if !cfg.AuditLogEnabled {
		return
	}
	entry := store.AuditLog{
		UID:       uid,
		Username:  username,
		Action:    action,
		Category:  category,
		TargetUID: targetUID,
		Detail:    detail,
		IP:        ip,
	}
	if cfg.AuditLogMaxEntries > 0 {
		_ = a.store().AddAuditLog(entry, cfg.AuditLogMaxEntries)
	} else {
		_ = a.store().AddAuditLog(entry, 10000)
	}
}

func (a *App) handleListAuditLogs(w http.ResponseWriter, r *http.Request, _ Params) {
	logs := a.store().ListAuditLogs()
	page := max(1, queryInt(r, "page", 1))
	perPage := clamp(queryInt(r, "per_page", 50), 1, 200)
	categoryFilter := strings.ToLower(r.URL.Query().Get("category"))
	actionFilter := strings.ToLower(r.URL.Query().Get("action"))
	uidFilter := r.URL.Query().Get("uid")
	search := strings.ToLower(r.URL.Query().Get("search"))

	filtered := make([]map[string]any, 0, len(logs))
	for _, log := range logs {
		if categoryFilter != "" && categoryFilter != "all" && log.Category != categoryFilter {
			continue
		}
		if actionFilter != "" && actionFilter != "all" && log.Action != actionFilter {
			continue
		}
		if uidFilter != "" {
			uid, _ := strconv.ParseInt(uidFilter, 10, 64)
			if uid > 0 && log.UID != uid {
				continue
			}
		}
		if search != "" {
			haystack := strings.ToLower(log.Username + " " + log.Action + " " + log.Category + " " + log.IP)
			if !strings.Contains(haystack, search) {
				continue
			}
		}
		filtered = append(filtered, auditLogDTO(log))
	}

	total := len(filtered)
	filtered = paginate(filtered, page, perPage)
	ok(w, "OK", map[string]any{"logs": filtered, "total": total, "page": page, "per_page": perPage})
}

func (a *App) handleDeleteAuditLog(w http.ResponseWriter, r *http.Request, params Params) {
	id, _ := strconv.ParseInt(params["id"], 10, 64)
	if id <= 0 {
		failWithCode(w, http.StatusBadRequest, ErrBadRequest, "无效的日志 ID")
		return
	}
	if err := a.store().DeleteAuditLog(id); err != nil {
		failWithCode(w, http.StatusNotFound, ErrNotFound, "日志不存在")
		return
	}
	ok(w, "已删除", nil)
}

func (a *App) handleClearAuditLogs(w http.ResponseWriter, r *http.Request, _ Params) {
	payload := decodeMap(r)
	if stringValue(payload, "confirm") != confirmClearAuditLogs {
		failWithCode(w, http.StatusBadRequest, ErrBadRequest, "需要确认短语 confirm="+confirmClearAuditLogs)
		return
	}
	if err := a.store().ClearAuditLogs(); err != nil {
		failWithCode(w, http.StatusInternalServerError, ErrInternal, "清空失败")
		return
	}
	ok(w, "审计日志已清空", nil)
}

// handlePruneAuditLogs 条件清理审计日志：支持按条数裁剪（max_entries）和按天数裁剪（retention_days），
// 两者可同时指定。需要确认短语。preserve_admin 控制是否保留管理员操作日志（仅对天数裁剪有效）。
func (a *App) handlePruneAuditLogs(w http.ResponseWriter, r *http.Request, _ Params) {
	payload := decodeMap(r)
	if stringValue(payload, "confirm") != confirmPruneAuditLogs {
		failWithCode(w, http.StatusBadRequest, ErrBadRequest, "需要确认短语 confirm="+confirmPruneAuditLogs)
		return
	}

	logs := []string{}

	// 按条数裁剪：保留最新 N 条
	if maxEntries := intValue(payload, "max_entries", 0); maxEntries > 0 {
		if err := a.store().PruneAuditLogs(maxEntries); err != nil {
			failWithCode(w, http.StatusInternalServerError, ErrInternal, "裁剪失败")
			return
		}
		logs = append(logs, fmt.Sprintf("保留最近 %d 条", maxEntries))
	}

	// 按天数裁剪：删除早于 retention_days 的记录
	if retentionDays := intValue(payload, "retention_days", 0); retentionDays > 0 {
		preserveAdmin := boolValue(payload, "preserve_admin", true)
		cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour).Unix()
		removed := a.store().PruneAuditLogsByAge(cutoff, preserveAdmin)
		logs = append(logs, fmt.Sprintf("删除 %d 天前 %d 条（保留管理员=%v）", retentionDays, removed, preserveAdmin))
	}

	if len(logs) == 0 {
		failWithCode(w, http.StatusBadRequest, ErrBadRequest, "请指定 max_entries 或 retention_days")
		return
	}

	ok(w, "审计日志已清理", map[string]any{
		"current": a.store().AuditLogCount(),
		"logs":    logs,
	})
}

func auditLogDTO(log store.AuditLog) map[string]any {
	return map[string]any{
		"id":         log.ID,
		"uid":        log.UID,
		"username":   log.Username,
		"action":     log.Action,
		"category":   log.Category,
		"target_uid": zeroNil(log.TargetUID),
		"detail":     log.Detail,
		"ip":         log.IP,
		"created_at": log.CreatedAt,
	}
}
