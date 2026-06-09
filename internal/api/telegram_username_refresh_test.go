package api

import (
	"testing"

	"github.com/prejudice-studio/twilight/internal/store"
)

// TestRefreshTelegramUsername 覆盖被动刷新的三种情形：用户名变化写库、空用户名
// 保留旧值、未绑定 telegram id 安全 no-op。
func TestRefreshTelegramUsername(t *testing.T) {
	app := newTestApp(t)
	u, err := app.store().CreateUser(store.User{Username: "alice", TelegramID: 555, TelegramUsername: "old_name"})
	if err != nil {
		t.Fatal(err)
	}

	// 新用户名（带 @ 前缀）→ 去前缀后写库。
	app.refreshTelegramUsername(555, "@New_Name")
	if got, _ := app.store().User(u.UID); got.TelegramUsername != "New_Name" {
		t.Fatalf("expected username refreshed to New_Name, got %q", got.TelegramUsername)
	}

	// 空用户名（对方删了 @username）→ 保留旧值，不清空。
	app.refreshTelegramUsername(555, "")
	if got, _ := app.store().User(u.UID); got.TelegramUsername != "New_Name" {
		t.Fatalf("empty username must not clear stored username, got %q", got.TelegramUsername)
	}

	// 相同用户名 → 无变化（不报错）。
	app.refreshTelegramUsername(555, "New_Name")
	if got, _ := app.store().User(u.UID); got.TelegramUsername != "New_Name" {
		t.Fatalf("idempotent refresh changed username, got %q", got.TelegramUsername)
	}

	// 未知 telegram id / 非法 id → 安全 no-op，不 panic。
	app.refreshTelegramUsername(999, "ghost")
	app.refreshTelegramUsername(0, "ignored")
}
