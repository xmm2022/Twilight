package api

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/prejudice-studio/twilight/internal/store"
)

func (a *App) RunTelegramBot(ctx context.Context) error {
	if !a.telegramAvailable() {
		return fmt.Errorf("Telegram mode is disabled or bot token is not configured")
	}
	me, err := a.telegramGetMe(ctx)
	if err != nil {
		return err
	}
	slog.Info("Telegram bot polling started", "username", me["username"])
	offset := int64(0)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		updates, err := a.telegramGetUpdates(ctx, offset)
		if err != nil {
			slog.Warn("Telegram getUpdates failed", "error", err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(3 * time.Second):
				continue
			}
		}
		for _, update := range updates {
			if id := numeric(update["update_id"]); id >= offset {
				offset = id + 1
			}
			a.handleTelegramUpdate(ctx, update)
		}
	}
}

func (a *App) telegramGetUpdates(ctx context.Context, offset int64) ([]map[string]any, error) {
	var result []map[string]any
	body := map[string]any{"timeout": 30, "allowed_updates": []string{"message", "chat_member", "my_chat_member"}}
	if offset > 0 {
		body["offset"] = offset
	}
	err := a.telegramPost(ctx, "getUpdates", body, &result)
	return result, err
}

func (a *App) handleTelegramUpdate(ctx context.Context, update map[string]any) {
	a.observeTelegramRoster(update)
	message, _ := update["message"].(map[string]any)
	if message == nil {
		return
	}
	text := strings.TrimSpace(asString(message["text"]))
	if text == "" {
		return
	}
	chat, _ := message["chat"].(map[string]any)
	from, _ := message["from"].(map[string]any)
	chatID := numeric(chat["id"])
	fromID := numeric(from["id"])
	if chatID == 0 {
		chatID = fromID
	}
	username := strings.TrimPrefix(asString(from["username"]), "@")
	fields := strings.Fields(text)
	command := strings.ToLower(fields[0])
	switch command {
	case "/start", "/help", "/twihelp":
		_ = a.telegramSendMessage(ctx, chatID, "Twilight Bot\n\n/bind <code> 绑定 Telegram\n/me 查看当前绑定\n/stats 查看服务统计")
	case "/me":
		a.telegramHandleMe(ctx, chatID, fromID)
	case "/stats":
		a.telegramHandleStats(ctx, chatID, fromID)
	case "/bind":
		if len(fields) < 2 {
			_ = a.telegramSendMessage(ctx, chatID, "请发送 /bind <绑定码>")
			return
		}
		a.telegramConfirmBindCode(ctx, chatID, fromID, username, fields[1])
	default:
		if len(command) >= 6 && len(command) <= 16 && !strings.HasPrefix(command, "/") {
			a.telegramConfirmBindCode(ctx, chatID, fromID, username, command)
		}
	}
}

func (a *App) observeTelegramRoster(update map[string]any) {
	if message, _ := update["message"].(map[string]any); message != nil {
		chat, _ := message["chat"].(map[string]any)
		from, _ := message["from"].(map[string]any)
		chatID := numeric(chat["id"])
		fromID := numeric(from["id"])
		if chatID != 0 && fromID > 0 && chatID != fromID {
			_ = a.store.UpsertTelegramRoster(fmt.Sprint(chatID), fromID, "member", boolish(from["is_bot"]))
		}
		return
	}
	for _, key := range []string{"chat_member", "my_chat_member"} {
		event, _ := update[key].(map[string]any)
		if event == nil {
			continue
		}
		chat, _ := event["chat"].(map[string]any)
		newMember, _ := event["new_chat_member"].(map[string]any)
		user, _ := newMember["user"].(map[string]any)
		chatID := numeric(chat["id"])
		userID := numeric(user["id"])
		status := strings.ToLower(asString(newMember["status"]))
		if chatID == 0 || userID <= 0 {
			return
		}
		if status == "left" || status == "kicked" {
			_ = a.store.MarkTelegramRosterLeft(fmt.Sprint(chatID), userID, status)
			return
		}
		_ = a.store.UpsertTelegramRoster(fmt.Sprint(chatID), userID, firstNonEmpty(status, "member"), boolish(user["is_bot"]))
		return
	}
}

func (a *App) telegramConfirmBindCode(ctx context.Context, chatID, telegramID int64, username, code string) {
	code = strings.ToUpper(strings.TrimSpace(code))
	bind, okBind := a.store.BindCode(code)
	if !okBind || bind.ExpiresAt < time.Now().Unix() {
		_ = a.telegramSendMessage(ctx, chatID, "绑定码无效或已过期，请在网页重新获取。")
		return
	}
	if existing, okUser := a.store.FindUserByTelegramID(telegramID); okUser && (bind.UID == 0 || existing.UID != bind.UID) {
		_ = a.telegramSendMessage(ctx, chatID, fmt.Sprintf("该 Telegram 已绑定到账号 %s。", existing.Username))
		return
	}
	bind.Confirmed = true
	bind.TelegramID = telegramID
	_ = a.store.UpsertBindCode(bind)
	if bind.UID != 0 {
		_, err := a.store.UpdateUser(bind.UID, func(u *store.User) error {
			u.TelegramID = telegramID
			u.TelegramUsername = username
			return nil
		})
		if err != nil {
			_ = a.telegramSendMessage(ctx, chatID, "绑定失败：系统用户不存在。")
			return
		}
	}
	_ = a.telegramSendMessage(ctx, chatID, "Telegram 绑定已确认，可以回到网页继续。")
}

func (a *App) telegramHandleMe(ctx context.Context, chatID, telegramID int64) {
	if u, okUser := a.store.FindUserByTelegramID(telegramID); okUser {
		_ = a.telegramSendMessage(ctx, chatID, fmt.Sprintf("当前绑定账号：%s (UID %d)", u.Username, u.UID))
		return
	}
	_ = a.telegramSendMessage(ctx, chatID, "当前 Telegram 尚未绑定 Twilight 账号。")
}

func (a *App) telegramHandleStats(ctx context.Context, chatID, telegramID int64) {
	if !a.telegramAdminID(telegramID) {
		_ = a.telegramSendMessage(ctx, chatID, "没有管理员权限。")
		return
	}
	users := a.store.ListUsers()
	active := 0
	embyBound := 0
	for _, u := range users {
		if u.Active {
			active++
		}
		if u.EmbyID != "" {
			embyBound++
		}
	}
	_ = a.telegramSendMessage(ctx, chatID, fmt.Sprintf("用户总数：%d\n活跃：%d\n已绑定 Emby：%d", len(users), active, embyBound))
}

func (a *App) telegramAdminID(telegramID int64) bool {
	for _, id := range a.cfg.TelegramAdminIDs {
		if id == telegramID {
			return true
		}
	}
	if u, okUser := a.store.FindUserByTelegramID(telegramID); okUser && u.Role == store.RoleAdmin {
		return true
	}
	return false
}
