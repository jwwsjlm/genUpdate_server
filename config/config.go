package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultUpdateInterval              = 5 * time.Minute
	DefaultServerPort                  = ":8090"
	DefaultReadTimeout                 = 15 * time.Second
	DefaultWriteTimeout                = 10 * time.Minute
	DefaultIdleTimeout                 = 60 * time.Second
	DefaultMaxConcurrentDownloads      = 64
	DefaultMaxConcurrentDownloadsPerIP = 8
	DefaultConfigFileName              = "config.json"
)

type Config struct {
	Port                        string        `json:"port"`
	UpdateDir                   string        `json:"updateDir"`
	ScanInterval                time.Duration `json:"scanInterval"`
	ReadTimeout                 time.Duration `json:"readTimeout"`
	WriteTimeout                time.Duration `json:"writeTimeout"`
	IdleTimeout                 time.Duration `json:"idleTimeout"`
	MaxConcurrentDownloads      int           `json:"maxConcurrentDownloads"`
	MaxConcurrentDownloadsPerIP int           `json:"maxConcurrentDownloadsPerIP"`
	AppTokens                   map[string]string
	WebPasswordHash             string
	WebSessionSecret            string
}

type FileConfig struct {
	Port                        *string           `json:"port"`
	UpdateDir                   *string           `json:"updateDir"`
	ScanIntervalSeconds         *int              `json:"scanIntervalSeconds"`
	ReadTimeoutSeconds          *int              `json:"readTimeoutSeconds"`
	WriteTimeoutSeconds         *int              `json:"writeTimeoutSeconds"`
	IdleTimeoutSeconds          *int              `json:"idleTimeoutSeconds"`
	MaxConcurrentDownloads      *int              `json:"maxConcurrentDownloads"`
	MaxConcurrentDownloadsPerIP *int              `json:"maxConcurrentDownloadsPerIP"`
	AppTokens                   map[string]string `json:"appTokens"`
	WebPasswordHash             *string           `json:"webPasswordHash"`
	WebSessionSecret            *string           `json:"webSessionSecret"`
}

func Load(workDir string) (Config, error) {
	cfg, err := defaultConfig(workDir)
	if err != nil {
		return Config{}, err
	}
	if err := applyConfigFile(&cfg, workDir); err != nil {
		return Config{}, err
	}
	if err := applyEnvOverrides(&cfg, workDir); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func defaultConfig(workDir string) (Config, error) {
	updateDir, err := filepath.Abs(filepath.Join(workDir, "update"))
	if err != nil {
		return Config{}, err
	}
	return Config{
		Port:                        DefaultServerPort,
		UpdateDir:                   updateDir,
		ScanInterval:                DefaultUpdateInterval,
		ReadTimeout:                 DefaultReadTimeout,
		WriteTimeout:                DefaultWriteTimeout,
		IdleTimeout:                 DefaultIdleTimeout,
		MaxConcurrentDownloads:      DefaultMaxConcurrentDownloads,
		MaxConcurrentDownloadsPerIP: DefaultMaxConcurrentDownloadsPerIP,
	}, nil
}

func GetServerPort() string {
	return normalizePort(os.Getenv("GENUPDATE_PORT"), DefaultServerPort)
}

func normalizePort(port, fallback string) string {
	port = strings.TrimSpace(port)
	if port == "" {
		return fallback
	}
	if port[0] != ':' {
		return ":" + port
	}
	return port
}

func applyConfigFile(cfg *Config, workDir string) error {
	configPath, explicit := configFilePath(workDir)
	if configPath == "" {
		return nil
	}
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) && !explicit {
			return nil
		}
		return fmt.Errorf("failed to stat config file %q: %w", configPath, err)
	}

	b, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %q: %w", configPath, err)
	}
	var fileCfg FileConfig
	if err := json.Unmarshal(b, &fileCfg); err != nil {
		return fmt.Errorf("failed to parse config file %q: %w", configPath, err)
	}
	return applyFileConfig(cfg, workDir, fileCfg)
}

func configFilePath(workDir string) (string, bool) {
	configPath := strings.TrimSpace(os.Getenv("GENUPDATE_CONFIG"))
	if configPath == "" {
		return filepath.Join(workDir, DefaultConfigFileName), false
	}
	return configPath, true
}

func applyFileConfig(cfg *Config, workDir string, fileCfg FileConfig) error {
	if fileCfg.Port != nil {
		cfg.Port = normalizePort(*fileCfg.Port, cfg.Port)
	}
	if fileCfg.UpdateDir != nil && strings.TrimSpace(*fileCfg.UpdateDir) != "" {
		updateDir := strings.TrimSpace(*fileCfg.UpdateDir)
		if !filepath.IsAbs(updateDir) {
			updateDir = filepath.Join(workDir, updateDir)
		}
		abs, err := filepath.Abs(updateDir)
		if err != nil {
			return err
		}
		cfg.UpdateDir = abs
	}
	if fileCfg.ScanIntervalSeconds != nil && *fileCfg.ScanIntervalSeconds > 0 {
		cfg.ScanInterval = time.Duration(*fileCfg.ScanIntervalSeconds) * time.Second
	}
	if fileCfg.ReadTimeoutSeconds != nil && *fileCfg.ReadTimeoutSeconds > 0 {
		cfg.ReadTimeout = time.Duration(*fileCfg.ReadTimeoutSeconds) * time.Second
	}
	if fileCfg.WriteTimeoutSeconds != nil && *fileCfg.WriteTimeoutSeconds > 0 {
		cfg.WriteTimeout = time.Duration(*fileCfg.WriteTimeoutSeconds) * time.Second
	}
	if fileCfg.IdleTimeoutSeconds != nil && *fileCfg.IdleTimeoutSeconds > 0 {
		cfg.IdleTimeout = time.Duration(*fileCfg.IdleTimeoutSeconds) * time.Second
	}
	if fileCfg.MaxConcurrentDownloads != nil && *fileCfg.MaxConcurrentDownloads > 0 {
		cfg.MaxConcurrentDownloads = *fileCfg.MaxConcurrentDownloads
	}
	if fileCfg.MaxConcurrentDownloadsPerIP != nil && *fileCfg.MaxConcurrentDownloadsPerIP > 0 {
		cfg.MaxConcurrentDownloadsPerIP = *fileCfg.MaxConcurrentDownloadsPerIP
	}
	if len(fileCfg.AppTokens) > 0 {
		cfg.AppTokens = cleanAppTokens(fileCfg.AppTokens)
	}
	if fileCfg.WebPasswordHash != nil {
		cfg.WebPasswordHash = strings.TrimSpace(*fileCfg.WebPasswordHash)
	}
	if fileCfg.WebSessionSecret != nil {
		cfg.WebSessionSecret = strings.TrimSpace(*fileCfg.WebSessionSecret)
	}
	return nil
}

func applyEnvOverrides(cfg *Config, workDir string) error {
	cfg.Port = GetServerPortWithFallback(cfg.Port)
	updateDir, err := getUpdateDir(workDir, cfg.UpdateDir)
	if err != nil {
		return err
	}
	cfg.UpdateDir = updateDir
	cfg.ScanInterval = GetDurationFromEnv("GENUPDATE_SCAN_INTERVAL_SECONDS", cfg.ScanInterval)
	cfg.ReadTimeout = GetDurationFromEnv("GENUPDATE_READ_TIMEOUT_SECONDS", cfg.ReadTimeout)
	cfg.WriteTimeout = GetDurationFromEnv("GENUPDATE_WRITE_TIMEOUT_SECONDS", cfg.WriteTimeout)
	cfg.IdleTimeout = GetDurationFromEnv("GENUPDATE_IDLE_TIMEOUT_SECONDS", cfg.IdleTimeout)
	cfg.MaxConcurrentDownloads = GetIntFromEnv("GENUPDATE_MAX_CONCURRENT_DOWNLOADS", cfg.MaxConcurrentDownloads)
	cfg.MaxConcurrentDownloadsPerIP = GetIntFromEnv("GENUPDATE_MAX_CONCURRENT_DOWNLOADS_PER_IP", cfg.MaxConcurrentDownloadsPerIP)
	if appTokens := GetAppTokensFromEnv("GENUPDATE_APP_TOKENS"); len(appTokens) > 0 {
		cfg.AppTokens = appTokens
	}
	if value := strings.TrimSpace(os.Getenv("GENUPDATE_WEB_PASSWORD_HASH")); value != "" {
		cfg.WebPasswordHash = value
	}
	if value := strings.TrimSpace(os.Getenv("GENUPDATE_WEB_SESSION_SECRET")); value != "" {
		cfg.WebSessionSecret = value
	}
	return nil
}

func GetServerPortWithFallback(fallback string) string {
	port := os.Getenv("GENUPDATE_PORT")
	if strings.TrimSpace(port) == "" {
		return fallback
	}
	return normalizePort(port, fallback)
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

func GetAppTokensFromEnv(name string) map[string]string {
	valueText := strings.TrimSpace(os.Getenv(name))
	if valueText == "" {
		return nil
	}

	return parseAppTokens(valueText)
}

func parseAppTokens(valueText string) map[string]string {
	tokens := make(map[string]string)
	for _, pair := range strings.Split(valueText, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		app, token, ok := strings.Cut(pair, "=")
		app = strings.TrimSpace(app)
		token = strings.TrimSpace(token)
		if !ok || app == "" || token == "" {
			continue
		}
		tokens[app] = token
	}
	if len(tokens) == 0 {
		return nil
	}
	return tokens
}

func cleanAppTokens(input map[string]string) map[string]string {
	tokens := make(map[string]string)
	for app, token := range input {
		app = strings.TrimSpace(app)
		token = strings.TrimSpace(token)
		if app == "" || token == "" {
			continue
		}
		tokens[app] = token
	}
	if len(tokens) == 0 {
		return nil
	}
	return tokens
}

func getUpdateDir(workDir, fallback string) (string, error) {
	updateDir := os.Getenv("GENUPDATE_UPDATE_DIR")
	if updateDir == "" {
		updateDir = fallback
	}
	return filepath.Abs(updateDir)
}
