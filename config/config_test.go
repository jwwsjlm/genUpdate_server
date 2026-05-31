package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetServerPort(t *testing.T) {
	t.Setenv("GENUPDATE_PORT", "")
	if got := GetServerPort(); got != DefaultServerPort {
		t.Fatalf("default port = %q, want %q", got, DefaultServerPort)
	}

	t.Setenv("GENUPDATE_PORT", "9000")
	if got := GetServerPort(); got != ":9000" {
		t.Fatalf("numeric port = %q, want :9000", got)
	}

	t.Setenv("GENUPDATE_PORT", ":9001")
	if got := GetServerPort(); got != ":9001" {
		t.Fatalf("prefixed port = %q, want :9001", got)
	}
}

func TestGetDurationFromEnv(t *testing.T) {
	t.Setenv("TEST_SECONDS", "")
	if got := GetDurationFromEnv("TEST_SECONDS", 7*time.Second); got != 7*time.Second {
		t.Fatalf("empty duration = %s, want 7s", got)
	}

	t.Setenv("TEST_SECONDS", "12")
	if got := GetDurationFromEnv("TEST_SECONDS", 7*time.Second); got != 12*time.Second {
		t.Fatalf("configured duration = %s, want 12s", got)
	}

	t.Setenv("TEST_SECONDS", "-1")
	if got := GetDurationFromEnv("TEST_SECONDS", 7*time.Second); got != 7*time.Second {
		t.Fatalf("invalid duration = %s, want 7s", got)
	}
}

func TestGetIntFromEnv(t *testing.T) {
	t.Setenv("TEST_INT", "")
	if got := GetIntFromEnv("TEST_INT", 3); got != 3 {
		t.Fatalf("empty int = %d, want 3", got)
	}

	t.Setenv("TEST_INT", "9")
	if got := GetIntFromEnv("TEST_INT", 3); got != 9 {
		t.Fatalf("configured int = %d, want 9", got)
	}

	t.Setenv("TEST_INT", "0")
	if got := GetIntFromEnv("TEST_INT", 3); got != 3 {
		t.Fatalf("invalid int = %d, want 3", got)
	}
}

func TestGetAppTokensFromEnv(t *testing.T) {
	t.Setenv("TEST_APP_TOKENS", "")
	if got := GetAppTokensFromEnv("TEST_APP_TOKENS"); got != nil {
		t.Fatalf("empty app tokens = %#v, want nil", got)
	}

	t.Setenv("TEST_APP_TOKENS", "cc=token-cc, bb = token-bb , broken, empty=")
	got := GetAppTokensFromEnv("TEST_APP_TOKENS")
	if got["cc"] != "token-cc" || got["bb"] != "token-bb" || len(got) != 2 {
		t.Fatalf("app tokens = %#v", got)
	}
}

func TestLoad(t *testing.T) {
	workDir := t.TempDir()

	t.Setenv("GENUPDATE_CONFIG", "")
	t.Setenv("GENUPDATE_PORT", "")
	t.Setenv("GENUPDATE_UPDATE_DIR", "")
	t.Setenv("GENUPDATE_MAX_CONCURRENT_DOWNLOADS", "5")
	t.Setenv("GENUPDATE_MAX_CONCURRENT_DOWNLOADS_PER_IP", "2")
	t.Setenv("GENUPDATE_APP_TOKENS", "cc=token-cc")
	cfg, err := Load(workDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.UpdateDir != filepath.Join(workDir, "update") {
		t.Fatalf("default update dir = %q", cfg.UpdateDir)
	}
	if cfg.MaxConcurrentDownloads != 5 {
		t.Fatalf("max concurrent downloads = %d, want 5", cfg.MaxConcurrentDownloads)
	}
	if cfg.MaxConcurrentDownloadsPerIP != 2 {
		t.Fatalf("max concurrent downloads per ip = %d, want 2", cfg.MaxConcurrentDownloadsPerIP)
	}
	if cfg.AppTokens["cc"] != "token-cc" {
		t.Fatalf("app token = %#v", cfg.AppTokens)
	}

	customDir := filepath.Join(workDir, "custom")
	t.Setenv("GENUPDATE_UPDATE_DIR", customDir)
	cfg, err = Load(workDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.UpdateDir != customDir {
		t.Fatalf("custom update dir = %q, want %q", cfg.UpdateDir, customDir)
	}
}

func TestLoadFromDefaultConfigFile(t *testing.T) {
	workDir := t.TempDir()
	writeConfigFile(t, filepath.Join(workDir, DefaultConfigFileName), `{
		"port": "9100",
		"updateDir": "updates",
		"scanIntervalSeconds": 42,
		"readTimeoutSeconds": 3,
		"writeTimeoutSeconds": 4,
		"idleTimeoutSeconds": 5,
		"maxConcurrentDownloads": 6,
		"maxConcurrentDownloadsPerIP": 2,
		"webPasswordHash": "$2a$10$example",
		"webSessionSecret": "session-secret",
		"appTokens": {
			"cc": "file-token"
		},
		"manifestSigningPrivateKey": "file-signing-key",
		"manifestSigningKeyID": "file-key-id"
	}`)
	t.Setenv("GENUPDATE_CONFIG", "")
	t.Setenv("GENUPDATE_PORT", "")
	t.Setenv("GENUPDATE_UPDATE_DIR", "")
	t.Setenv("GENUPDATE_APP_TOKENS", "")

	cfg, err := Load(workDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != ":9100" {
		t.Fatalf("port = %q, want :9100", cfg.Port)
	}
	if cfg.UpdateDir != filepath.Join(workDir, "updates") {
		t.Fatalf("update dir = %q, want %q", cfg.UpdateDir, filepath.Join(workDir, "updates"))
	}
	if cfg.ScanInterval != 42*time.Second || cfg.ReadTimeout != 3*time.Second || cfg.WriteTimeout != 4*time.Second || cfg.IdleTimeout != 5*time.Second {
		t.Fatalf("durations = scan %s read %s write %s idle %s", cfg.ScanInterval, cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout)
	}
	if cfg.MaxConcurrentDownloads != 6 {
		t.Fatalf("max concurrent downloads = %d, want 6", cfg.MaxConcurrentDownloads)
	}
	if cfg.MaxConcurrentDownloadsPerIP != 2 {
		t.Fatalf("max concurrent downloads per ip = %d, want 2", cfg.MaxConcurrentDownloadsPerIP)
	}
	if cfg.AppTokens["cc"] != "file-token" {
		t.Fatalf("app tokens = %#v", cfg.AppTokens)
	}
	if cfg.WebPasswordHash != "$2a$10$example" || cfg.WebSessionSecret != "session-secret" {
		t.Fatalf("web auth config = hash %q secret %q", cfg.WebPasswordHash, cfg.WebSessionSecret)
	}
	if cfg.ManifestSigningPrivateKey != "file-signing-key" || cfg.ManifestSigningKeyID != "file-key-id" {
		t.Fatalf("manifest signing config = key %q id %q", cfg.ManifestSigningPrivateKey, cfg.ManifestSigningKeyID)
	}
}

func TestLoadEnvOverridesConfigFile(t *testing.T) {
	workDir := t.TempDir()
	configPath := filepath.Join(workDir, "custom-config.json")
	writeConfigFile(t, configPath, `{
		"port": "9100",
		"maxConcurrentDownloads": 6,
		"maxConcurrentDownloadsPerIP": 2,
		"appTokens": {
			"cc": "file-token"
		}
	}`)
	t.Setenv("GENUPDATE_CONFIG", configPath)
	t.Setenv("GENUPDATE_PORT", "9200")
	t.Setenv("GENUPDATE_MAX_CONCURRENT_DOWNLOADS", "9")
	t.Setenv("GENUPDATE_MAX_CONCURRENT_DOWNLOADS_PER_IP", "4")
	t.Setenv("GENUPDATE_APP_TOKENS", "cc=env-token")
	t.Setenv("GENUPDATE_WEB_PASSWORD_HASH", "$2a$10$env")
	t.Setenv("GENUPDATE_WEB_SESSION_SECRET", "env-session-secret")
	t.Setenv("GENUPDATE_MANIFEST_SIGNING_PRIVATE_KEY", "env-signing-key")
	t.Setenv("GENUPDATE_MANIFEST_SIGNING_KEY_ID", "env-key-id")

	cfg, err := Load(workDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != ":9200" {
		t.Fatalf("port = %q, want :9200", cfg.Port)
	}
	if cfg.MaxConcurrentDownloads != 9 {
		t.Fatalf("max concurrent downloads = %d, want 9", cfg.MaxConcurrentDownloads)
	}
	if cfg.MaxConcurrentDownloadsPerIP != 4 {
		t.Fatalf("max concurrent downloads per ip = %d, want 4", cfg.MaxConcurrentDownloadsPerIP)
	}
	if cfg.AppTokens["cc"] != "env-token" {
		t.Fatalf("app tokens = %#v", cfg.AppTokens)
	}
	if cfg.WebPasswordHash != "$2a$10$env" || cfg.WebSessionSecret != "env-session-secret" {
		t.Fatalf("web auth config = hash %q secret %q", cfg.WebPasswordHash, cfg.WebSessionSecret)
	}
	if cfg.ManifestSigningPrivateKey != "env-signing-key" || cfg.ManifestSigningKeyID != "env-key-id" {
		t.Fatalf("manifest signing config = key %q id %q", cfg.ManifestSigningPrivateKey, cfg.ManifestSigningKeyID)
	}
}

func TestLoadExplicitMissingConfigFile(t *testing.T) {
	workDir := t.TempDir()
	t.Setenv("GENUPDATE_CONFIG", filepath.Join(workDir, "missing.json"))

	if _, err := Load(workDir); err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
}

func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
