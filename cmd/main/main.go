package main

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jwwsjlm/genUpdate_server/auth"
	"github.com/jwwsjlm/genUpdate_server/config"
	"github.com/jwwsjlm/genUpdate_server/fileutils"
	"github.com/jwwsjlm/genUpdate_server/route"
	"go.uber.org/zap"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	setupLogger()
	defer closeLogger()

	workDir, err := os.Getwd()
	if err != nil {
		auth.Panicf("failed to get working directory: %v", err)
	}

	cfg, err := config.Load(workDir)
	if err != nil {
		auth.Panicf("failed to load config: %v", err)
	}

	initUpdateList(cfg.UpdateDir)

	ticker := time.NewTicker(cfg.ScanInterval)
	defer ticker.Stop()
	go periodicUpdate(ticker, cfg.UpdateDir)

	auth.Infof("server starting on %s, update dir: %s, scan interval: %s", cfg.Port, cfg.UpdateDir, cfg.ScanInterval)
	startServer(cfg)
}

func setupLogger() {
	auth.InitLogger(zap.InfoLevel)
}

func closeLogger() {
	if err := auth.Logger.Sync(); err != nil {
		auth.Error("failed to sync logger", zap.Error(err))
	}
}

func initUpdateList(dir string) {
	ignorePath := filepath.Join(dir, ".ignore")
	if err := fileutils.InitListUpdate(ignorePath, dir); err != nil {
		auth.Panicf("failed to read update files: %v", err)
	}
	logFileListSummary()
}

func logFileListSummary() {
	jsonData, err := fileutils.GetJSONText()
	if err != nil {
		auth.Errorf("failed to build file list JSON: %v", err)
		return
	}
	auth.Infof("file list JSON generated, length: %d bytes", len(jsonData))
}

func periodicUpdate(ticker *time.Ticker, dir string) {
	ignorePath := filepath.Join(dir, ".ignore")
	for range ticker.C {
		if err := fileutils.InitListUpdate(ignorePath, dir); err != nil {
			auth.Errorf("failed to refresh update list: %v", err)
		}
	}
}

func startServer(cfg config.Config) {
	server := &http.Server{
		Addr: cfg.Port,
		Handler: route.SetupRouterWithOptions(route.Options{
			UpdateDir:                   cfg.UpdateDir,
			MaxConcurrentDownloads:      cfg.MaxConcurrentDownloads,
			MaxConcurrentDownloadsPerIP: cfg.MaxConcurrentDownloadsPerIP,
			AppTokens:                   cfg.AppTokens,
			Build: route.BuildInfo{
				Version:   version,
				Commit:    commit,
				BuildTime: buildTime,
			},
		}),
		ReadHeaderTimeout: cfg.ReadTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		auth.Panicf("server failed: %v", err)
	}
}
