package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadUsesMySQLDefaults(t *testing.T) {
	clearConfigEnv(t)

	cfg := Load()

	if strings.Contains(cfg.DSN(), "sqlite") {
		t.Fatalf("expected mysql DSN, got %q", cfg.DSN())
	}
	if !strings.Contains(cfg.DSN(), "root:123456@tcp(localhost:3306)/noteweb") {
		t.Fatalf("expected default mysql DSN, got %q", cfg.DSN())
	}
}

func TestLoadReadsDotEnvFile(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	t.Chdir(dir)

	dotEnv := strings.Join([]string{
		"MYSQL_HOST=127.0.0.1",
		"MYSQL_PORT=3310",
		"MYSQL_USER=xxladmin",
		"MYSQL_PASSWORD=XXLadmin_2021!",
		"MYSQL_DATABASE=noteweb",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(dotEnv), 0600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg := Load()

	want := "xxladmin:XXLadmin_2021!@tcp(127.0.0.1:3310)/noteweb?charset=utf8mb4&parseTime=True&loc=Local"
	if cfg.DSN() != want {
		t.Fatalf("expected DSN %q, got %q", want, cfg.DSN())
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"APP_NAME",
		"ENV",
		"DEBUG",
		"PORT",
		"SECRET_KEY",
		"ACCESS_TOKEN_EXPIRE_MINUTES",
		"MYSQL_HOST",
		"MYSQL_PORT",
		"MYSQL_USER",
		"MYSQL_PASSWORD",
		"MYSQL_DATABASE",
		"MAX_UPLOAD_SIZE",
	} {
		t.Setenv(key, "")
	}
}
