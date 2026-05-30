package route

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jwwsjlm/genUpdate_server/auth"
	"github.com/jwwsjlm/genUpdate_server/fileutils"
	"go.uber.org/zap"
)

const (
	defaultMaxConcurrentDownloads = 64
	manifestCacheMaxAge           = "60s"
)

var errInvalidDownloadPath = errors.New("invalid download path")

type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

type Options struct {
	UpdateDir              string
	MaxConcurrentDownloads int
	Build                  BuildInfo
	AppTokens              map[string]string
}

type downloadLimiter struct {
	sem chan struct{}
}

func newDownloadLimiter(limit int) *downloadLimiter {
	if limit <= 0 {
		limit = defaultMaxConcurrentDownloads
	}
	return &downloadLimiter{sem: make(chan struct{}, limit)}
}

func (l *downloadLimiter) acquire() bool {
	select {
	case l.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (l *downloadLimiter) release() {
	select {
	case <-l.sem:
	default:
	}
}

func SetupRouter(updateDir string) *gin.Engine {
	return SetupRouterWithOptions(Options{UpdateDir: updateDir})
}

func SetupRouterWithOptions(opts Options) *gin.Engine {
	r := gin.New()
	r.Use(securityHeaders())
	r.Use(ginLogger(auth.Logger))
	r.Use(gin.Recovery())
	gin.SetMode(gin.ReleaseMode)

	state := routerState{
		updateDir: opts.UpdateDir,
		limiter:   newDownloadLimiter(opts.MaxConcurrentDownloads),
		build:     opts.Build,
		appTokens: opts.AppTokens,
	}

	r.GET("/healthz", state.healthz)
	r.GET("/version", state.version)
	r.GET("/", state.index)
	r.GET("/api/apps", state.apps)
	r.GET("/updateList/:filename", state.getUpdateList)
	r.GET("/download/*filepath", state.download)
	r.HEAD("/download/*filepath", state.download)

	return r
}

type routerState struct {
	updateDir string
	limiter   *downloadLimiter
	build     BuildInfo
	appTokens map[string]string
}

func (s routerState) healthz(c *gin.Context) {
	if s.updateDir == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"ret": "error", "error": "update dir not configured"})
		return
	}
	if info, err := os.Stat(s.updateDir); err != nil || !info.IsDir() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"ret": "error", "error": "update dir unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": "ok", "status": "healthy"})
}

func (s routerState) version(c *gin.Context) {
	jsonText, err := fileutils.GetJSONText()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ret": "error", "error": "failed to load version info"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ret":           "ok",
		"version":       s.build.Version,
		"commit":        s.build.Commit,
		"buildTime":     s.build.BuildTime,
		"fileListBytes": len(jsonText),
		"cacheMaxAge":   manifestCacheMaxAge,
	})
}

func (s routerState) apps(c *gin.Context) {
	lists := fileutils.GetAllLists()
	lists, ok := s.filterAuthorizedLists(c, lists)
	if !ok {
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusUnauthorized, gin.H{"ret": "error", "error": "unauthorized"})
		return
	}

	totalFiles := 0
	totalBytes := int64(0)
	for _, app := range lists {
		totalFiles += len(app.Files)
		for _, f := range app.Files {
			totalBytes += f.Size
		}
	}

	if len(s.appTokens) == 0 {
		c.Header("Cache-Control", "public, max-age=60")
	} else {
		c.Header("Cache-Control", "private, no-store")
	}
	c.JSON(http.StatusOK, gin.H{
		"ret":        "ok",
		"apps":       lists,
		"totalApps":  len(lists),
		"totalFiles": totalFiles,
		"totalBytes": totalBytes,
	})
}

func (s routerState) index(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, indexHTML)
}

func (s routerState) getUpdateList(c *gin.Context) {
	filename := c.Param("filename")
	if !s.authorizeApp(c, filename) {
		c.JSON(http.StatusNotFound, gin.H{"ret": "error", "error": "software not found"})
		return
	}
	if f, ok := fileutils.GetList(filename); ok {
		if len(s.appTokens) == 0 {
			c.Header("Cache-Control", "public, max-age=60")
		} else {
			c.Header("Cache-Control", "private, no-store")
		}
		c.JSON(http.StatusOK, gin.H{"ret": "ok", "appList": f})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"ret": "error", "error": "software not found"})
}

func (s routerState) download(c *gin.Context) {
	if !s.limiter.acquire() {
		c.JSON(http.StatusTooManyRequests, gin.H{"ret": "error", "error": "too many concurrent downloads"})
		return
	}
	defer s.limiter.release()

	filePath, cleanPath, err := resolveDownloadPath(s.updateDir, c.Param("filepath"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ret": "error", "error": "invalid file path"})
		return
	}
	if !fileutils.HasFilePath(cleanPath) {
		c.JSON(http.StatusNotFound, gin.H{"ret": "error", "error": "file not found"})
		return
	}
	if !s.authorizeApp(c, appNameFromCleanPath(cleanPath)) {
		c.JSON(http.StatusNotFound, gin.H{"ret": "error", "error": "file not found"})
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"ret": "error", "error": "file not found"})
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || info.IsDir() {
		c.JSON(http.StatusNotFound, gin.H{"ret": "error", "error": "file not found"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(cleanPath))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Accept-Ranges", "bytes")
	c.Header("ETag", weakETag(info))
	http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), file)
}

func weakETag(info os.FileInfo) string {
	h := sha256.New()
	_, _ = io.WriteString(h, info.Name())
	_, _ = io.WriteString(h, fmt.Sprintf(":%d:%d", info.Size(), info.ModTime().UnixNano()))
	return `W/"` + hex.EncodeToString(h.Sum(nil)) + `"`
}

func (s routerState) filterAuthorizedLists(c *gin.Context, lists []fileutils.FileList) ([]fileutils.FileList, bool) {
	if len(s.appTokens) == 0 {
		return lists, true
	}

	filtered := make([]fileutils.FileList, 0, len(lists))
	for _, list := range lists {
		if s.authorizeApp(c, list.FileName) {
			filtered = append(filtered, list)
		}
	}
	return filtered, len(filtered) > 0
}

func (s routerState) authorizeApp(c *gin.Context, appName string) bool {
	if len(s.appTokens) == 0 {
		return true
	}
	expectedToken, ok := s.appTokens[appName]
	if !ok || expectedToken == "" {
		return false
	}
	token := requestToken(c)
	if token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1
}

func requestToken(c *gin.Context) string {
	const bearerPrefix = "bearer "
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if len(authHeader) >= len(bearerPrefix) && strings.EqualFold(authHeader[:len(bearerPrefix)], bearerPrefix) {
		if token := strings.TrimSpace(authHeader[len(bearerPrefix):]); token != "" {
			return token
		}
	}
	return strings.TrimSpace(c.GetHeader("X-Update-Token"))
}

func appNameFromCleanPath(cleanPath string) string {
	appName, _, _ := strings.Cut(cleanPath, "/")
	return appName
}

func resolveDownloadPath(rootDir, requestPath string) (string, string, error) {
	relPath := strings.TrimPrefix(requestPath, "/")
	if relPath == "" {
		return "", "", errInvalidDownloadPath
	}

	cleanPath := filepath.Clean(filepath.FromSlash(relPath))
	if cleanPath == "." || filepath.IsAbs(cleanPath) || cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return "", "", errInvalidDownloadPath
	}

	rootAbs, err := filepath.Abs(rootDir)
	if err != nil {
		return "", "", err
	}
	filePath := filepath.Join(rootAbs, cleanPath)
	fileAbs, err := filepath.Abs(filePath)
	if err != nil {
		return "", "", err
	}

	relToRoot, err := filepath.Rel(rootAbs, fileAbs)
	if err != nil || relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
		return "", "", errInvalidDownloadPath
	}

	return fileAbs, filepath.ToSlash(cleanPath), nil
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; base-uri 'none'; frame-ancestors 'none'")
		c.Next()
	}
}

func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if logger == nil {
			return
		}

		for _, e := range c.Errors {
			logger.Error("request error", zap.String("error", e.Error()))
		}

		logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}
