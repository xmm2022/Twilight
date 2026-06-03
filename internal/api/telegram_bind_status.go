package api

import (
	"net/http"
	"strings"

	"github.com/prejudice-studio/twilight/internal/store"
)

type registerTelegramBindCodeState struct {
	Code             string
	Status           string
	ErrorCode        ErrCode
	HTTPStatus       int
	Message          string
	Bind             store.BindCode
	Confirmed        bool
	Invalid          bool
	Terminal         bool
	ExpiresIn        int64
	TelegramID       int64
	TelegramUsername string
}

func (a *App) registerTelegramBindCodeState(code string, now int64, cleanupExpired bool) registerTelegramBindCodeState {
	code = strings.ToUpper(strings.TrimSpace(code))
	if !telegramBindCodePattern.MatchString(code) {
		return registerTelegramBindCodeState{Code: code, Status: "invalid_format", ErrorCode: ErrTGBindCodeFormat, HTTPStatus: http.StatusBadRequest, Message: "Telegram 绑定码格式不正确", Invalid: true, Terminal: true}
	}
	bind, okBind := a.store().BindCode(code)
	if !okBind {
		return registerTelegramBindCodeState{Code: code, Status: "not_found", ErrorCode: ErrTGBindCodeNotFound, HTTPStatus: http.StatusBadRequest, Message: "绑定码不存在", Invalid: true, Terminal: true}
	}
	if bind.ExpiresAt <= now {
		if cleanupExpired {
			_ = a.store().DeleteBindCode(code)
		}
		return registerTelegramBindCodeState{Code: code, Status: "expired", ErrorCode: ErrTGBindCodeExpired, HTTPStatus: http.StatusBadRequest, Message: "绑定码无效或已过期", Bind: bind, Invalid: true, Terminal: true}
	}
	state := registerTelegramBindCodeState{Code: code, Bind: bind, ExpiresIn: bind.ExpiresAt - now, TelegramID: bind.TelegramID, TelegramUsername: bind.TelegramUsername}
	if bind.Scene != "register" || bind.UID != 0 {
		state.Status = "wrong_scene"
		state.ErrorCode = ErrTGBindCodeSceneBad
		state.HTTPStatus = http.StatusBadRequest
		state.Message = "绑定码场景无效"
		state.Invalid = true
		state.Terminal = true
		return state
	}
	if !bind.Confirmed || bind.TelegramID == 0 {
		state.Status = "pending"
		state.ErrorCode = ErrTGBindCodeNotConfirm
		state.HTTPStatus = http.StatusBadRequest
		state.Message = "绑定码尚未在 Telegram 中确认"
		state.Confirmed = false
		state.Terminal = false
		return state
	}
	if existing, okUser := a.store().FindUserByTelegramID(bind.TelegramID); okUser {
		state.Status = "telegram_taken"
		state.ErrorCode = ErrTGAlreadyBound
		state.HTTPStatus = http.StatusConflict
		state.Message = "该 Telegram 已绑定到账号 " + existing.Username
		state.Confirmed = true
		state.Invalid = true
		state.Terminal = true
		return state
	}
	state.Status = "confirmed"
	state.Message = "绑定码已确认"
	state.Confirmed = true
	state.Terminal = true
	return state
}

func (s registerTelegramBindCodeState) response() map[string]any {
	data := map[string]any{
		"code":           s.Code,
		"status":         s.Status,
		"confirmed":      s.Confirmed,
		"invalid":        s.Invalid,
		"terminal":       s.Terminal,
		"message":        s.Message,
		"telegram_bound": s.Status == "confirmed",
	}
	if s.ErrorCode != "" {
		data["error_code"] = s.ErrorCode
	}
	if s.ExpiresIn > 0 {
		data["expires_in"] = s.ExpiresIn
	}
	if s.TelegramID != 0 {
		data["telegram_id"] = s.TelegramID
	}
	if strings.TrimSpace(s.TelegramUsername) != "" {
		data["telegram_username"] = s.TelegramUsername
	}
	return data
}
