package main

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jwwsjlm/genUpdate_server/auth"
	"github.com/jwwsjlm/genUpdate_server/fileutils"
	"github.com/jwwsjlm/genUpdate_server/route"
	"go.uber.org/zap"
)

const (
	defaultUpdateInterval = 5 * time.Minute
	defaultServerPort     = ":8090"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	setupLogger()
	defer closeLogger()

	dir, err := os.Getwd()
	if err != nil {
		auth.Panicf("获取当前目录失败: %s", err.Error())
	}

	updateDir := filepath.Join(dir, "update")
	initUpdateList(updateDir)

	interval := getUpdateInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	go periodicUpdate(ticker, updateDir)

	startServer(getServerPort())
}

func setupLogger() {
	auth.InitLogger(zap.InfoLevel)
}

func closeLogger() {
	if err := auth.Logger.Sync(); err != nil {
		auth.Error("关闭 zap log sync 出错", zap.Error(err))
	}
}

func initUpdateList(dir string) {
	ignorePath := filepath.Join(dir, ".ignore")
	if err := fileutils.InitListUpdate(ignorePath, dir); err != nil {
		auth.Panicf("读取文件失败: %v", err)
	}
	logFileListSummary()
}

func logFileListSummary() {
	jsonData, err := fileutils.GetJSONText()
	if err != nil {
		auth.Errorf("生成文件列表 JSON 失败: %v", err)
		return
	}
	auth.Infof("文件 JSON 生成成功，长度: %d bytes", len(jsonData))
}

func periodicUpdate(ticker *time.Ticker, dir string) {
	ignorePath := filepath.Join(dir, ".ignore")
	for range ticker.C {
		if err := fileutils.InitListUpdate(ignorePath, dir); err != nil {
			auth.Errorf("刷新列表失败: %v", err)
		}
	}
}

func startServer(port string) {
	r := route.SetupRouter()
	if err := r.Run(port); err != nil {
		auth.Panicf("Gin 启动失败: %v", err)
	}
}

func getServerPort() string {
	port := os.Getenv("GENUPDATE_PORT")
	if port == "" {
		return defaultServerPort
	}
	if port[0] != ':' {
		return ":" + port
	}
	return port
}

func getUpdateInterval() time.Duration {
	secondsText := os.Getenv("GENUPDATE_SCAN_INTERVAL_SECONDS")
	if secondsText == "" {
		return defaultUpdateInterval
	}
	seconds, err := strconv.Atoi(secondsText)
	if err != nil || seconds <= 0 {
		return defaultUpdateInterval
	}
	return time.Duration(seconds) * time.Second
}
