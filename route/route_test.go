package route

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jwwsjlm/genUpdate_server/fileutils"
)

func TestResolveDownloadPath(t *testing.T) {
	root := t.TempDir()

	tests := []struct {
		name        string
		requestPath string
		wantClean   string
		wantErr     bool
	}{
		{name: "valid nested path", requestPath: "/app/data/file.txt", wantClean: "app/data/file.txt"},
		{name: "empty path", requestPath: "/", wantErr: true},
		{name: "parent traversal", requestPath: "/../secret.txt", wantErr: true},
		{name: "nested parent traversal", requestPath: "/app/../../secret.txt", wantErr: true},
		{name: "filename containing dots", requestPath: "/app/v1..2.txt", wantClean: "app/v1..2.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotClean, err := resolveDownloadPath(root, tt.requestPath)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveDownloadPath() error = %v", err)
			}
			if gotClean != tt.wantClean {
				t.Fatalf("clean path = %q, want %q", gotClean, tt.wantClean)
			}
			wantPath := filepath.Join(root, filepath.FromSlash(tt.wantClean))
			if gotPath != wantPath {
				t.Fatalf("file path = %q, want %q", gotPath, wantPath)
			}
		})
	}
}

func TestDownloadSupportsHeadAndRangeRequests(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	root := createDownloadFixture(t)
	router := SetupRouter(root)

	t.Run("range request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/download/app/file.txt", nil)
		req.Header.Set("Range", "bytes=2-5")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusPartialContent {
			t.Fatalf("status = %d, want %d, body = %q", rec.Code, http.StatusPartialContent, rec.Body.String())
		}
		if got := rec.Body.String(); got != "2345" {
			t.Fatalf("body = %q, want 2345", got)
		}
		if got := rec.Header().Get("Content-Range"); got != "bytes 2-5/10" {
			t.Fatalf("Content-Range = %q, want bytes 2-5/10", got)
		}
		if got := rec.Header().Get("Accept-Ranges"); got != "bytes" {
			t.Fatalf("Accept-Ranges = %q, want bytes", got)
		}
		if got := rec.Header().Get("ETag"); got == "" {
			t.Fatalf("ETag header is empty")
		}
	})

	t.Run("head request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodHead, "/download/app/file.txt", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if rec.Body.Len() != 0 {
			t.Fatalf("HEAD body length = %d, want 0", rec.Body.Len())
		}
		if got := rec.Header().Get("Content-Length"); got != "10" {
			t.Fatalf("Content-Length = %q, want 10", got)
		}
	})
}

func TestHealthAndVersion(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	root := createDownloadFixture(t)
	router := SetupRouterWithOptions(Options{
		UpdateDir: root,
		Build:     BuildInfo{Version: "1.2.3", Commit: "abc123", BuildTime: "2026-05-30T00:00:00Z"},
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", rec.Code)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/version", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("version status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !containsAll(body, "1.2.3", "abc123", "fileListBytes") {
		t.Fatalf("version body missing fields: %s", body)
	}
}

func TestWebIndexAndAppsAPI(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	root := createDownloadFixture(t)
	if err := fileutils.InitListUpdate(filepath.Join(root, ".ignore"), root); err != nil {
		t.Fatalf("InitListUpdate() error = %v", err)
	}
	router := SetupRouter(root)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("index status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !containsAll(body, "GenUpdate", "/api/apps", "更新中心") {
		t.Fatalf("index body missing expected content: %s", body)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/apps", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("apps status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !containsAll(body, `"ret":"ok"`, `"totalApps":1`, `"fileName":"app"`, `"downloadURL":"/download/app/file.txt"`) {
		t.Fatalf("apps body missing expected content: %s", body)
	}
}

func TestDownloadLimiter(t *testing.T) {
	limiter := newDownloadLimiter(1)
	if !limiter.acquire() {
		t.Fatalf("first acquire failed")
	}
	if limiter.acquire() {
		t.Fatalf("second acquire succeeded, want limited")
	}
	limiter.release()
	if !limiter.acquire() {
		t.Fatalf("acquire after release failed")
	}
}

func createDownloadFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	appDir := filepath.Join(root, "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	filePath := filepath.Join(appDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("0123456789"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return root
}

func containsAll(text string, values ...string) bool {
	for _, value := range values {
		if !strings.Contains(text, value) {
			return false
		}
	}
	return true
}
