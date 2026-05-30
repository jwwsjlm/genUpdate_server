package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	DefaultUpdateInterval         = 5 * time.Minute
	DefaultServerPort             = ":8090"
	DefaultReadTimeout            = 15 * time.Second
	DefaultWriteTimeout           = 10 * time.Minute
	DefaultIdleTimeout            = 60 * time.Second
	DefaultMaxConcurrentDownloads = 64
)

type Config struct {
	Port                   string        `json:"port"`
	UpdateDir              string        `json:"updateDir"`
	ScanInterval           time.Duration `json:"scanInterval"`
	ReadTimeout            time.Duration `json:"readTimeout"`
	WriteTimeout           time.Duration `json:"writeTimeout"`
	IdleTimeout            time.Duration `json:"idleTimeout"`
	MaxConcurrentDownloads int           `json:"maxConcurrentDownloads"`
}

func Load(workDir string) (Config, error) {
	updateDir, err := getUpdateDir(workDir)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:                   GetServerPort(),
		UpdateDir:              updateDir,
		ScanInterval:           GetDurationFromEnv("GENUPDATE_SCAN_INTERVAL_SECONDS", DefaultUpdateInterval),
		ReadTimeout:            GetDurationFromEnv("GENUPDATE_READ_TIMEOUT_SECONDS", DefaultReadTimeout),
		WriteTimeout:           GetDurationFromEnv("GENUPDATE_WRITE_TIMEOUT_SECONDS", DefaultWriteTimeout),
		IdleTimeout:            GetDurationFromEnv("GENUPDATE_IDLE_TIMEOUT_SECONDS", DefaultIdleTimeout),
		MaxConcurrentDownloads: GetIntFromEnv("GENUPDATE_MAX_CONCURRENT_DOWNLOADS", DefaultMaxConcurrentDownloads),
	}, nil
}

func GetServerPort() string {
	port := os.Getenv("GENUPDATE_PORT")
	if port == "" {
		return DefaultServerPort
	}
	if port[0] != ':' {
		return ":" + port
	}
	return port
}

func GetDurationFromEnv(name string, fallback time.Duration) time.Duration {
	secondsText := os.Getenv(name)
	if secondsText == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(secondsText)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func GetIntFromEnv(name string, fallback int) int {
	valueText := os.Getenv(name)
	if valueText == "" {
		return fallback
	}
	value, err := strconv.Atoi(valueText)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func getUpdateDir(workDir string) (string, error) {
	updateDir := os.Getenv("GENUPDATE_UPDATE_DIR")
	if updateDir == "" {
		updateDir = filepath.Join(workDir, "update")
	}
	return filepath.Abs(updateDir)
}
