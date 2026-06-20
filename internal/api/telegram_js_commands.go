package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
)

const telegramJSPrefix = "js:"

func (a *App) telegramHandleCustomCommand(ctx context.Context, command string, c telegramCommandCtx, privateChat bool) bool {
	reply, ok := a.telegramCustomCommandReply(command)
	if !ok {
		return false
	}
	trimmed := strings.TrimSpace(reply)
	if !strings.HasPrefix(strings.ToLower(trimmed), telegramJSPrefix) {
		_ = a.telegramSendMessage(ctx, c.ChatID, a.telegramRenderText(reply))
		return true
	}

	text, logs, err := a.telegramRunJSCustomCommand(strings.TrimSpace(trimmed[len(telegramJSPrefix):]), c, privateChat)
	user, _ := a.store().FindUserByTelegramID(c.FromID)
	detail := map[string]any{"command": telegramCommand(command), "ok": err == nil, "private_chat": privateChat}
	if len(logs) > 0 {
		detail["logs"] = logs
	}
	a.auditEntryIP("telegram", user.UID, user.Username, "telegram_js_command_execute", "system", user.UID, detail)
	if err != nil {
		_ = a.telegramSendMessage(ctx, c.ChatID, "自定义指令执行失败，请联系管理员查看安全审计。")
		return true
	}
	if strings.TrimSpace(text) == "" {
		text = "自定义指令已执行。"
	}
	_ = a.telegramSendMessage(ctx, c.ChatID, a.telegramRenderText(text))
	return true
}

func (a *App) telegramRunJSCustomCommand(code string, c telegramCommandCtx, privateChat bool) (string, []string, error) {
	result := validateDeveloperJSCommand(code)
	if ok, _ := result["ok"].(bool); !ok {
		return "", nil, fmt.Errorf("developer js command rejected: %v", result["errors"])
	}

	user, _ := a.store().FindUserByTelegramID(c.FromID)
	vm := goja.New()
	replies := make([]string, 0, 4)
	logs := make([]string, 0, 8)
	_ = vm.Set("ctx", map[string]any{
		"private_chat": privateChat,
		"command_time": time.Now().Unix(),
	})
	_ = vm.Set("args", c.Args)
	_ = vm.Set("user", map[string]any{
		"uid":      user.UID,
		"username": user.Username,
		"role":     user.Role,
		"active":   user.Active,
		"has_emby": strings.TrimSpace(user.EmbyID) != "",
	})
	_ = vm.Set("reply", func(call goja.FunctionCall) goja.Value {
		if len(replies) < 4 {
			replies = append(replies, call.Argument(0).String())
		}
		return goja.Undefined()
	})
	_ = vm.Set("log", func(call goja.FunctionCall) goja.Value {
		if len(logs) < 8 {
			logs = append(logs, call.Argument(0).String())
		}
		return goja.Undefined()
	})

	timer := time.AfterFunc(200*time.Millisecond, func() {
		vm.Interrupt("execution timeout")
	})
	defer timer.Stop()
	if _, err := vm.RunString(code); err != nil {
		return "", logs, err
	}
	return strings.Join(replies, "\n"), logs, nil
}
