package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTOMLAndEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `[Global]
redis_url = "redis://localhost:6379/2"

[API]
host = "127.0.0.1"
port = 5050
cors_origins = ["http://localhost:3000", "https://example.com"]
max_upload_size = 1234
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TWILIGHT_API_PORT", "6060")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RedisURL != "redis://localhost:6379/2" {
		t.Fatalf("unexpected redis url: %q", cfg.RedisURL)
	}
	if cfg.Host != "127.0.0.1" || cfg.Port != 6060 {
		t.Fatalf("unexpected host/port: %s/%d", cfg.Host, cfg.Port)
	}
	if len(cfg.CORSOrigins) != 2 || cfg.MaxUploadSize != 1234 {
		t.Fatalf("unexpected cors/upload config: %#v %d", cfg.CORSOrigins, cfg.MaxUploadSize)
	}
}

func TestLoadMultilineArraysAndPostgresConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `[Global]
databases_dir = "db"

[Database]
driver = "postgres"
postgres_host = "db.local"
postgres_port = 5433
postgres_user = "twilight"
postgres_password = "secret"
postgres_database = "twilight_prod"
postgres_sslmode = "require"

[Emby]
emby_url_list = [
  "Direct : http://127.0.0.1:8096/",
  "Relay : https://emby.example.com/",
]
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseDriver != "postgres" || cfg.PostgresPort != 5433 {
		t.Fatalf("unexpected database config: %#v", cfg)
	}
	if cfg.PostgresDSN() == "" || cfg.PostgresSSLMode != "require" {
		t.Fatalf("expected postgres dsn, got %q", cfg.PostgresDSN())
	}
	if len(cfg.EmbyURLList) != 2 || cfg.EmbyURLList[0].Name != "Direct" {
		t.Fatalf("unexpected emby lines: %#v", cfg.EmbyURLList)
	}
}
