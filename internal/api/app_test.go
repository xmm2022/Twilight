package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prejudice-studio/twilight/internal/config"
	"github.com/prejudice-studio/twilight/internal/store"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	app, err := New(config.Config{
		AppName:                      "Twilight Test",
		Version:                      "test",
		Host:                         "127.0.0.1",
		Port:                         0,
		DatabaseDir:                  dir,
		DatabaseBackupDir:            filepath.Join(dir, "backups"),
		StateFile:                    filepath.Join(dir, "state.json"),
		UploadDir:                    filepath.Join(dir, "uploads"),
		MaxUploadSize:                1024 * 1024,
		CORSOrigins:                  []string{"http://localhost:3000"},
		AllowCredential:              true,
		SessionCookie:                "twilight_session",
		SessionTTL:                   time.Hour,
		CookieSameSite:               "lax",
		MediaRequestEnabled:          true,
		MaxConcurrentRequestsPerUser: 3,
		InviteEnabled:                true,
		InviteMaxDepth:               3,
		InviteLimit:                  10,
		InviteRootUserLimit:          -1,
		InviteDefaultDays:            30,
		PermanentInviteMaxDays:       365,
		MaxDevices:                   5,
		MaxStreams:                   2,
	}, st)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func TestAuthFlowAndCSRFMitigation(t *testing.T) {
	app := newTestApp(t)

	register := doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"admin","password":"admin123456"}`, nil)
	if register.Code != http.StatusCreated {
		t.Fatalf("register status = %d body=%s", register.Code, register.Body.String())
	}

	login := doJSON(app, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"admin123456"}`, nil)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", login.Code, login.Body.String())
	}
	cookie := findCookie(login.Result().Cookies(), "twilight_session")
	if cookie == nil || !cookie.HttpOnly {
		t.Fatalf("expected httponly session cookie, got %#v", cookie)
	}

	me := doJSON(app, http.MethodGet, "/api/v1/users/me", ``, []*http.Cookie{cookie})
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d body=%s", me.Code, me.Body.String())
	}
	if me.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing security header")
	}

	blocked := doJSON(app, http.MethodPut, "/api/v1/users/me", `{"email":"a@example.com"}`, []*http.Cookie{cookie})
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("csrf status = %d body=%s", blocked.Code, blocked.Body.String())
	}

	allowed := doJSONWithHeaders(app, http.MethodPut, "/api/v1/users/me", `{"email":"a@example.com"}`, []*http.Cookie{cookie}, map[string]string{"X-Twilight-Client": "webui"})
	if allowed.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", allowed.Code, allowed.Body.String())
	}
}

func TestAPIKeyFlow(t *testing.T) {
	app := newTestApp(t)
	_ = doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"admin","password":"admin123456"}`, nil)
	login := doJSON(app, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"admin123456"}`, nil)
	cookie := findCookie(login.Result().Cookies(), "twilight_session")

	created := doJSONWithHeaders(app, http.MethodPost, "/api/v1/users/me/apikeys", `{"name":"ci","rate_limit":50}`, []*http.Cookie{cookie}, map[string]string{"X-Twilight-Client": "webui"})
	if created.Code != http.StatusOK {
		t.Fatalf("create key status = %d body=%s", created.Code, created.Body.String())
	}
	var env envelope
	if err := json.Unmarshal(created.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	data := env.Data.(map[string]any)
	key, _ := data["key"].(string)
	if !strings.HasPrefix(key, "key-") {
		t.Fatalf("expected plaintext key once, got %q", key)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apikey/info", nil)
	req.Header.Set("X-API-Key", key)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("apikey info status = %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFrontendRouteCompatibilityDoesNot404(t *testing.T) {
	app := newTestApp(t)
	routes := []struct{ method, path string }{
		{http.MethodGet, "/api/v1/system/info"},
		{http.MethodGet, "/api/v1/system/health"},
		{http.MethodPost, "/api/v1/auth/login"},
		{http.MethodGet, "/api/v1/users/me"},
		{http.MethodGet, "/api/v1/admin/users"},
		{http.MethodGet, "/api/v1/media/search?q=x&source=tmdb"},
		{http.MethodPost, "/api/v1/media/request"},
		{http.MethodGet, "/api/v1/admin/scheduler/jobs"},
		{http.MethodGet, "/api/v1/announcements"},
		{http.MethodGet, "/api/v1/invite/config"},
		{http.MethodGet, "/api/v1/signin/config"},
		{http.MethodGet, "/api/v1/demo/bootstrap"},
		{http.MethodGet, "/api/v1/apikey/info"},
	}
	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, nil)
		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)
		if rr.Code == http.StatusNotFound || rr.Code == http.StatusMethodNotAllowed {
			t.Fatalf("%s %s returned %d", route.method, route.path, rr.Code)
		}
	}
}

func TestUploadRejectsNonImage(t *testing.T) {
	app := newTestApp(t)
	_ = doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"admin","password":"admin123456"}`, nil)
	login := doJSON(app, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"admin123456"}`, nil)
	cookie := findCookie(login.Result().Cookies(), "twilight_session")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "note.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("not an image"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/avatar/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Twilight-Client", "webui")
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("upload status = %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRegcodeInviteMediaAndSecurityFlows(t *testing.T) {
	app := newTestApp(t)
	_ = doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"admin","password":"admin123456"}`, nil)
	adminLogin := doJSON(app, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"admin123456"}`, nil)
	adminCookie := findCookie(adminLogin.Result().Cookies(), "twilight_session")

	_ = doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"user","password":"user123456"}`, nil)
	userLogin := doJSON(app, http.MethodPost, "/api/v1/auth/login", `{"username":"user","password":"user123456"}`, nil)
	userCookie := findCookie(userLogin.Result().Cookies(), "twilight_session")
	user, _ := app.store.FindUserByUsername("user")
	_, _ = app.store.UpdateUser(user.UID, func(u *store.User) error { u.TelegramID = 12345; return nil })

	createdCodes := doJSONWithHeaders(app, http.MethodPost, "/api/v1/admin/regcodes", `{"type":2,"days":15,"count":1,"random_algorithm":"hex20"}`, []*http.Cookie{adminCookie}, map[string]string{"X-Twilight-Client": "webui"})
	if createdCodes.Code != http.StatusOK {
		t.Fatalf("create regcode status=%d body=%s", createdCodes.Code, createdCodes.Body.String())
	}
	var codeEnv envelope
	if err := json.Unmarshal(createdCodes.Body.Bytes(), &codeEnv); err != nil {
		t.Fatal(err)
	}
	code := codeEnv.Data.(map[string]any)["codes"].([]any)[0].(string)

	preview := doJSONWithHeaders(app, http.MethodPost, "/api/v1/users/me/use-code", `{"reg_code":"`+code+`","check_only":true}`, []*http.Cookie{userCookie}, map[string]string{"X-Twilight-Client": "webui"})
	if preview.Code != http.StatusOK || !strings.Contains(preview.Body.String(), "续期") {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	used := doJSONWithHeaders(app, http.MethodPost, "/api/v1/users/me/use-code", `{"reg_code":"`+code+`"}`, []*http.Cookie{userCookie}, map[string]string{"X-Twilight-Client": "webui"})
	if used.Code != http.StatusOK {
		t.Fatalf("use code status=%d body=%s", used.Code, used.Body.String())
	}

	invite := doJSONWithHeaders(app, http.MethodPost, "/api/v1/invite/codes", `{"days":7}`, []*http.Cookie{adminCookie}, map[string]string{"X-Twilight-Client": "webui"})
	if invite.Code != http.StatusCreated {
		t.Fatalf("invite status=%d body=%s", invite.Code, invite.Body.String())
	}
	forest := doJSONWithHeaders(app, http.MethodGet, "/api/v1/admin/invite/tree", ``, []*http.Cookie{adminCookie}, nil)
	if forest.Code != http.StatusOK || !strings.Contains(forest.Body.String(), "nodes") {
		t.Fatalf("forest status=%d body=%s", forest.Code, forest.Body.String())
	}

	media := doJSONWithHeaders(app, http.MethodPost, "/api/v1/media/request", `{"source":"tmdb","media_id":550,"title":"Fight Club","media_type":"movie"}`, []*http.Cookie{userCookie}, map[string]string{"X-Twilight-Client": "webui"})
	if media.Code != http.StatusCreated || !strings.Contains(media.Body.String(), "require_key") {
		t.Fatalf("media status=%d body=%s", media.Code, media.Body.String())
	}
	adminReqs := doJSONWithHeaders(app, http.MethodGet, "/api/v1/admin/media-requests?status=all", ``, []*http.Cookie{adminCookie}, nil)
	if adminReqs.Code != http.StatusOK || !strings.Contains(adminReqs.Body.String(), "Fight Club") {
		t.Fatalf("admin reqs status=%d body=%s", adminReqs.Code, adminReqs.Body.String())
	}

	blocked := doJSONWithHeaders(app, http.MethodPost, "/api/v1/security/ip/blacklist", `{"ip":"203.0.113.9","reason":"test"}`, []*http.Cookie{adminCookie}, map[string]string{"X-Twilight-Client": "webui"})
	if blocked.Code != http.StatusOK {
		t.Fatalf("blacklist status=%d body=%s", blocked.Code, blocked.Body.String())
	}
}

func TestInventoryCheckUsesEmbyProviderAndSeasons(t *testing.T) {
	app := newTestApp(t)
	emby := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		q := r.URL.Query()
		if q.Get("AnyProviderIdEquals") == "Tmdb.42" {
			_, _ = w.Write([]byte(`{"Items":[{"Id":"series1","Name":"Example Show","Type":"Series","ProductionYear":2024,"ProviderIds":{"Tmdb":"42"}}],"TotalRecordCount":1}`))
			return
		}
		if q.Get("ParentId") == "series1" {
			_, _ = w.Write([]byte(`{"Items":[{"Id":"s1","Name":"Season 1","Type":"Season","IndexNumber":1},{"Id":"s2","Name":"Season 2","Type":"Season","IndexNumber":2}],"TotalRecordCount":2}`))
			return
		}
		_, _ = w.Write([]byte(`{"Items":[],"TotalRecordCount":0}`))
	}))
	defer emby.Close()
	app.cfg.EmbyURL = emby.URL

	_ = doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"admin","password":"admin123456"}`, nil)
	login := doJSON(app, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"admin123456"}`, nil)
	cookie := findCookie(login.Result().Cookies(), "twilight_session")

	missingSeason := doJSONWithHeaders(app, http.MethodPost, "/api/v1/media/inventory/check", `{"source":"tmdb","media_id":42,"media_type":"tv","season":3}`, []*http.Cookie{cookie}, map[string]string{"X-Twilight-Client": "webui"})
	if missingSeason.Code != http.StatusOK || !strings.Contains(missingSeason.Body.String(), `"exists":false`) || !strings.Contains(missingSeason.Body.String(), `"seasons_available":[1,2]`) {
		t.Fatalf("missing season status=%d body=%s", missingSeason.Code, missingSeason.Body.String())
	}
	existingSeason := doJSONWithHeaders(app, http.MethodPost, "/api/v1/media/inventory/check", `{"source":"tmdb","media_id":42,"media_type":"tv","season":2}`, []*http.Cookie{cookie}, map[string]string{"X-Twilight-Client": "webui"})
	if existingSeason.Code != http.StatusOK || !strings.Contains(existingSeason.Body.String(), `"exists":true`) || !strings.Contains(existingSeason.Body.String(), `"season_requested":2`) {
		t.Fatalf("existing season status=%d body=%s", existingSeason.Code, existingSeason.Body.String())
	}
}

func TestSystemUpdateRejectsUnsafeRepoURL(t *testing.T) {
	app := newTestApp(t)
	_ = doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"admin","password":"admin123456"}`, nil)
	login := doJSON(app, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"admin123456"}`, nil)
	cookie := findCookie(login.Result().Cookies(), "twilight_session")
	resp := doJSONWithHeaders(app, http.MethodPost, "/api/v1/system/admin/update", `{"repo_url":"https://user:pass@example.com/repo.git","branch":"main"}`, []*http.Cookie{cookie}, map[string]string{"X-Twilight-Client": "webui"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("unsafe update URL status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestTelegramBindConfirmRequiresInternalSecret(t *testing.T) {
	app := newTestApp(t)
	app.cfg.BotInternalSecret = "test-secret"
	now := time.Now().Unix()
	if err := app.store.UpsertBindCode(store.BindCode{Code: "ABCDEFGH", Scene: "register", CreatedAt: now, ExpiresAt: now + 60}); err != nil {
		t.Fatal(err)
	}
	blocked := doJSON(app, http.MethodPost, "/api/v1/users/me/telegram/bind-confirm", `{"code":"ABCDEFGH","telegram_id":42}`, nil)
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("bind confirm without secret = %d body=%s", blocked.Code, blocked.Body.String())
	}
	allowed := doJSONWithHeaders(app, http.MethodPost, "/api/v1/users/me/telegram/bind-confirm", `{"code":"ABCDEFGH","telegram_id":42}`, nil, map[string]string{"X-Internal-Secret": "test-secret"})
	if allowed.Code != http.StatusOK {
		t.Fatalf("bind confirm with secret = %d body=%s", allowed.Code, allowed.Body.String())
	}
}

func TestBangumiWebhookRequiresSecretWhenEnabled(t *testing.T) {
	app := newTestApp(t)
	app.cfg.BangumiEnabled = true
	blocked := doJSON(app, http.MethodPost, "/api/v1/emby/bangumi/webhook", `{"Event":"PlaybackStopped"}`, nil)
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("webhook without configured secret = %d body=%s", blocked.Code, blocked.Body.String())
	}
	app.cfg.BangumiWebhookSecret = "webhook-secret"
	allowed := doJSON(app, http.MethodPost, "/api/v1/emby/bangumi/webhook?token=webhook-secret", `{"Event":"PlaybackStopped"}`, nil)
	if allowed.Code != http.StatusOK {
		t.Fatalf("webhook with secret = %d body=%s", allowed.Code, allowed.Body.String())
	}
}

func TestDatabaseAdminBackupRestoreAndAuth(t *testing.T) {
	app := newTestApp(t)
	_ = doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"admin","password":"admin123456"}`, nil)
	adminLogin := doJSON(app, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"admin123456"}`, nil)
	adminCookie := findCookie(adminLogin.Result().Cookies(), "twilight_session")
	_ = doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"user","password":"user123456"}`, nil)
	userLogin := doJSON(app, http.MethodPost, "/api/v1/auth/login", `{"username":"user","password":"user123456"}`, nil)
	userCookie := findCookie(userLogin.Result().Cookies(), "twilight_session")

	unauth := doJSON(app, http.MethodGet, "/api/v1/system/admin/database/status", ``, nil)
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("database status unauth = %d", unauth.Code)
	}
	forbidden := doJSON(app, http.MethodGet, "/api/v1/system/admin/database/status", ``, []*http.Cookie{userCookie})
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("database status user = %d body=%s", forbidden.Code, forbidden.Body.String())
	}
	backup := doJSONWithHeaders(app, http.MethodPost, "/api/v1/system/admin/database/backup", `{}`, []*http.Cookie{adminCookie}, map[string]string{"X-Twilight-Client": "webui"})
	if backup.Code != http.StatusOK {
		t.Fatalf("backup status=%d body=%s", backup.Code, backup.Body.String())
	}
	var env envelope
	if err := json.Unmarshal(backup.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	backupData := env.Data.(map[string]any)["backup"].(map[string]any)
	backupName := backupData["name"].(string)

	_ = doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"extra","password":"extra123456"}`, nil)
	if _, ok := app.store.FindUserByUsername("extra"); !ok {
		t.Fatal("expected extra user before restore")
	}
	restore := doJSONWithHeaders(app, http.MethodPost, "/api/v1/system/admin/database/restore", `{"name":"`+backupName+`"}`, []*http.Cookie{adminCookie}, map[string]string{"X-Twilight-Client": "webui"})
	if restore.Code != http.StatusOK {
		t.Fatalf("restore status=%d body=%s", restore.Code, restore.Body.String())
	}
	if _, ok := app.store.FindUserByUsername("extra"); ok {
		t.Fatal("restore did not replace state")
	}
	traversal := doJSONWithHeaders(app, http.MethodPost, "/api/v1/system/admin/database/restore", `{"name":"../state.json"}`, []*http.Cookie{adminCookie}, map[string]string{"X-Twilight-Client": "webui"})
	if traversal.Code != http.StatusBadRequest {
		t.Fatalf("restore traversal status=%d body=%s", traversal.Code, traversal.Body.String())
	}
	migrate := doJSONWithHeaders(app, http.MethodPost, "/api/v1/system/admin/database/migrate", `{"target_driver":"json","dry_run":true}`, []*http.Cookie{adminCookie}, map[string]string{"X-Twilight-Client": "webui"})
	if migrate.Code != http.StatusOK || !strings.Contains(migrate.Body.String(), `"dry_run":true`) {
		t.Fatalf("migrate dry-run status=%d body=%s", migrate.Code, migrate.Body.String())
	}
}

func TestTelegramRosterStatsUsesObservedMembers(t *testing.T) {
	app := newTestApp(t)
	app.cfg.TelegramGroupIDs = []string{"-1001"}
	if err := app.store.UpsertTelegramRoster("-1001", 100, "member", false); err != nil {
		t.Fatal(err)
	}
	if err := app.store.UpsertTelegramRoster("-1001", 200, "member", false); err != nil {
		t.Fatal(err)
	}
	if err := app.store.UpsertTelegramRoster("-1001", 300, "member", true); err != nil {
		t.Fatal(err)
	}
	_ = doJSON(app, http.MethodPost, "/api/v1/users/register", `{"username":"admin","password":"admin123456"}`, nil)
	admin, _ := app.store.FindUserByUsername("admin")
	_, _ = app.store.UpdateUser(admin.UID, func(u *store.User) error { u.TelegramID = 100; return nil })
	login := doJSON(app, http.MethodPost, "/api/v1/auth/login", `{"username":"admin","password":"admin123456"}`, nil)
	cookie := findCookie(login.Result().Cookies(), "twilight_session")
	resp := doJSONWithHeaders(app, http.MethodGet, "/api/v1/admin/telegram/roster/stats", ``, []*http.Cookie{cookie}, nil)
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), `"bound":1`) || !strings.Contains(resp.Body.String(), `"unbound":1`) || !strings.Contains(resp.Body.String(), `"bots":1`) {
		t.Fatalf("roster stats status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func doJSON(app *App, method, path, body string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	return doJSONWithHeaders(app, method, path, body, cookies, nil)
}

func doJSONWithHeaders(app *App, method, path, body string, cookies []*http.Cookie, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	return rr
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
