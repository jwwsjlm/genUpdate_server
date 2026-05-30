package config

import (
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

func TestLoad(t *testing.T) {
	workDir := t.TempDir()

	t.Setenv("GENUPDATE_UPDATE_DIR", "")
	t.Setenv("GENUPDATE_MAX_CONCURRENT_DOWNLOADS", "5")
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
