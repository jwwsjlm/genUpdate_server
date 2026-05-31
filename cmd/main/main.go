package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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
	"golang.org/x/crypto/bcrypt"
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
	if handleCLI(os.Args) {
		return
	}

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

func handleCLI(args []string) bool {
	if len(args) < 2 {
		return false
	}
	switch args[1] {
	case "hash-password":
		if len(args) < 3 {
			_, _ = fmt.Fprintln(os.Stderr, "usage: genupdate-server hash-password <password>")
			os.Exit(2)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(args[2]), bcrypt.DefaultCost)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to hash password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(hash))
		return true
	case "generate-signing-key":
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to generate signing key: %v\n", err)
			os.Exit(1)
		}
		publicKey := privateKey.Public().(ed25519.PublicKey)
		keyIDHash := sha256.Sum256(publicKey)
		fmt.Println("GENUPDATE_MANIFEST_SIGNING_PRIVATE_KEY=" + base64.RawURLEncoding.EncodeToString(privateKey.Seed()))
		fmt.Println("GENUPDATE_MANIFEST_SIGNING_PUBLIC_KEY=" + base64.RawURLEncoding.EncodeToString(publicKey))
		fmt.Println("GENUPDATE_MANIFEST_SIGNING_KEY_ID=" + base64.RawURLEncoding.EncodeToString(keyIDHash[:8]))
		return true
	default:
		return false
	}
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
			WebPasswordHash:             cfg.WebPasswordHash,
			WebSessionSecret:            cfg.WebSessionSecret,
			ManifestSigningPrivateKey:   cfg.ManifestSigningPrivateKey,
			ManifestSigningKeyID:        cfg.ManifestSigningKeyID,
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
