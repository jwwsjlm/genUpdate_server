package main

import (
	"os"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/jwwsjlm/genUpdate_srver/auth"
	"github.com/jwwsjlm/genUpdate_srver/db"
	"github.com/jwwsjlm/genUpdate_srver/fileutils"
	"github.com/jwwsjlm/genUpdate_srver/route"
	"go.uber.org/zap"
)

const (
	updateInterval = 5 * time.Minute
	serverPort     = ":8090"
)

// init 初始化函数
func init() {
	gin.SetMode(gin.ReleaseMode)
}
func feclose() {
	if err := db.Close(); err != nil {
		auth.Error("RoseDB 关闭失败: %w", zap.Error(err))
	}
	if err := auth.Logger.Sync(); err != nil {
		auth.Error("关闭 zap log sync 出错: %s", zap.Error(err))
	}
}

// main 主函数
func main() {
	// 初始化日志
	setupLogger()
	defer feclose()
	// 获取当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		auth.Panicf("获取当前目录失败: %s", err.Error())
	}

	// 初始化数据库
	initDatabase(dir)
	// 设置定时器用于定期更新文件列表
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	// 初始化更新列表
	updateDir := dir + "/update/"
	initUpdateList(updateDir)

	// 启动定期更新文件列表的 goroutine
	go periodicUpdate(ticker, updateDir)

	// 设置并启动路由
	startServer()
}

// setupLogger 设置日志
func setupLogger() {
	auth.InitLogger(zap.InfoLevel)
}

// initDatabase 初始化数据库
func initDatabase(dir string) {
	dbPath := dir + "/tmp/roseDb_basic"
	if err := db.NewRoseDb(dbPath); err != nil {
		auth.Panicf("RoseDB 数据库加载失败: %v", err)
	}
}

// initUpdateList 初始化更新列表
func initUpdateList(dir string) {
	if err := fileutils.InitListUpdate("update/.ignore", dir); err != nil {
		auth.Panicf("读取文件失败: %v", err)
	}
	logFileListJson()
}

// logFileListJson 记录文件列表 JSON
func logFileListJson() {
	jsonData, err := json.Marshal(fileutils.FileListJson)
	if err != nil {
		auth.Errorf("生成文件列表 JSON 失败: %v", err)
	} else {
		auth.Infof("文件 JSON 生成: %s", jsonData)
	}
}

// periodicUpdate 定期更新文件列表
func periodicUpdate(ticker *time.Ticker, dir string) {
	for range ticker.C {
		if err := fileutils.InitListUpdate("update/.ignore", dir); err != nil {
			auth.Errorf("刷新列表失败: %v", err)
		}
	}
}

// startServer 启动服务器
func startServer() {
	r := route.SetupRouter()
	if err := r.Run(serverPort); err != nil {
		auth.Panicf("Gin 启动失败: %v", err)
	}
}
