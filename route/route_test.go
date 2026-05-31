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
	"golang.org/x/crypto/bcrypt"
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
	if err := fileutils.InitListUpdate(filepath.Join(root, ".ignore"), root); err != nil {
		t.Fatalf("InitListUpdate() error = %v", err)
	}
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

func TestDownloadOnlyServesManifestFiles(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	root := createDownloadFixture(t)
	if err := os.WriteFile(filepath.Join(root, "app", ".env"), []byte("secret"), 0o644); err != nil {
		t.Fatalf("WriteFile(.env) error = %v", err)
	}
	privateDir := filepath.Join(root, "app", ".private")
	if err := os.MkdirAll(privateDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(.private) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(privateDir, "token.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatalf("WriteFile(token.txt) error = %v", err)
	}
	if err := fileutils.InitListUpdate(filepath.Join(root, ".ignore"), root); err != nil {
		t.Fatalf("InitListUpdate() error = %v", err)
	}
	router := SetupRouter(root)

	tests := []string{
		"/download/jsonBody.json",
		"/download/manifest-cache.json",
		"/download/app/.env",
		"/download/app/.private/token.txt",
	}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404, body = %q", rec.Code, rec.Body.String())
			}
		})
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/apps", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("apps status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); strings.Contains(body, ".env") || strings.Contains(body, ".private") {
		t.Fatalf("apps body leaked hidden files: %s", body)
	}
}

func TestAppTokenScopesListsAndDownloads(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "cc", "cc.exe"), "cc")
	writeTestFile(t, filepath.Join(root, "bb", "bb.exe"), "bb")
	if err := fileutils.InitListUpdate(filepath.Join(root, ".ignore"), root); err != nil {
		t.Fatalf("InitListUpdate() error = %v", err)
	}
	router := SetupRouterWithOptions(Options{
		UpdateDir: root,
		AppTokens: map[string]string{"cc": "cc-token", "bb": "bb-token"},
	})

	t.Run("apps requires token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/apps", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
		if body := rec.Body.String(); strings.Contains(body, "cc") || strings.Contains(body, "bb") {
			t.Fatalf("unauthorized apps response leaked names: %s", body)
		}
	})

	t.Run("apps only returns authorized app", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
		req.Header.Set("Authorization", "Bearer cc-token")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if !strings.Contains(body, `"fileName":"cc"`) || strings.Contains(body, `"fileName":"bb"`) {
			t.Fatalf("apps body scoped incorrectly: %s", body)
		}
	})

	t.Run("update list hides other apps", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/updateList/bb", nil)
		req.Header.Set("X-Update-Token", "cc-token")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404, body = %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("download hides other apps", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/download/bb/bb.exe", nil)
		req.Header.Set("Authorization", "Bearer cc-token")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404, body = %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("download allows matching app token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/download/cc/cc.exe", nil)
		req.Header.Set("Authorization", "Bearer cc-token")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %s", rec.Code, rec.Body.String())
		}
		if got := rec.Body.String(); got != "cc" {
			t.Fatalf("body = %q, want cc", got)
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
	if body := rec.Body.String(); !containsAll(body, "GenUpdate", "/api/apps", "更新中心", "generateTokenBtn") {
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

func TestWebPasswordProtectsIndexAndAppsAPI(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	root := createDownloadFixture(t)
	if err := fileutils.InitListUpdate(filepath.Join(root, ".ignore"), root); err != nil {
		t.Fatalf("InitListUpdate() error = %v", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	router := SetupRouterWithOptions(Options{
		UpdateDir:        root,
		WebPasswordHash:  string(hash),
		WebSessionSecret: "test-session-secret",
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("login page status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !containsAll(body, "loginPassword", "/api/web-login") || strings.Contains(body, "/api/apps") {
		t.Fatalf("login page body unexpected: %s", body)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/apps", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("apps status = %d, want 401", rec.Code)
	}
	if body := rec.Body.String(); strings.Contains(body, `"fileName":"app"`) {
		t.Fatalf("unauthorized apps response leaked app list: %s", body)
	}

	badLogin := httptest.NewRequest(http.MethodPost, "/api/web-login", strings.NewReader("password=wrong"))
	badLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, badLogin)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad login status = %d, want 401", rec.Code)
	}

	goodLogin := httptest.NewRequest(http.MethodPost, "/api/web-login", strings.NewReader("password=secret"))
	goodLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, goodLogin)
	if rec.Code != http.StatusOK {
		t.Fatalf("good login status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("login did not set cookie")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("authorized apps status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `"fileName":"app"`) {
		t.Fatalf("authorized apps body missing app list: %s", body)
	}
}

func TestWebLoginRateLimit(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	root := createDownloadFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	router := SetupRouterWithOptions(Options{
		UpdateDir:        root,
		WebPasswordHash:  string(hash),
		WebSessionSecret: "test-session-secret",
	})

	for i := 0; i < defaultWebLoginAttemptsPerMinute; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/web-login", strings.NewReader("password=wrong"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want 401", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/web-login", strings.NewReader("password=wrong"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("rate limited status = %d, want 429", rec.Code)
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

func TestClientDownloadLimiter(t *testing.T) {
	limiter := newClientDownloadLimiter(1)
	if !limiter.acquire("192.0.2.1") {
		t.Fatalf("first acquire failed")
	}
	if limiter.acquire("192.0.2.1") {
		t.Fatalf("second acquire for same client succeeded, want limited")
	}
	if !limiter.acquire("192.0.2.2") {
		t.Fatalf("different client acquire failed")
	}
	limiter.release("192.0.2.1")
	if !limiter.acquire("192.0.2.1") {
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
	writeTestFile(t, filepath.Join(appDir, "file.txt"), "0123456789")
	return root
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func containsAll(text string, values ...string) bool {
	for _, value := range values {
		if !strings.Contains(text, value) {
			return false
		}
	}
	return true
}
