package route

import (
	"crypto/sha256"
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
	r.Use(ginLogger(auth.Logger))
	r.Use(gin.Recovery())
	gin.SetMode(gin.ReleaseMode)

	state := routerState{
		updateDir: opts.UpdateDir,
		limiter:   newDownloadLimiter(opts.MaxConcurrentDownloads),
		build:     opts.Build,
	}

	r.GET("/healthz", state.healthz)
	r.GET("/version", state.version)
	r.GET("/", state.index)
	r.GET("/api/apps", state.apps)
	r.GET("/updateList/:filename", getUpdateList)
	r.GET("/download/*filepath", state.download)
	r.HEAD("/download/*filepath", state.download)

	return r
}

type routerState struct {
	updateDir string
	limiter   *downloadLimiter
	build     BuildInfo
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
	totalFiles := 0
	totalBytes := int64(0)
	for _, app := range lists {
		totalFiles += len(app.Files)
		for _, f := range app.Files {
			totalBytes += f.Size
		}
	}

	c.Header("Cache-Control", "public, max-age=60")
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

func getUpdateList(c *gin.Context) {
	filename := c.Param("filename")
	if f, ok := fileutils.GetList(filename); ok {
		c.Header("Cache-Control", "public, max-age=60")
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
