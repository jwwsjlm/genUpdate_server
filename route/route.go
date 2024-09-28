package route

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/gin-gonic/gin"
	"github.com/jwwsjlm/genUpdate_server/auth"
	"github.com/jwwsjlm/genUpdate_server/db"
	"github.com/jwwsjlm/genUpdate_server/fileutils"
	"go.uber.org/zap"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(ginLogger(auth.Logger))
	gin.SetMode(gin.ReleaseMode)

	// 设置路由
	r.GET("updateList/:filename", getUpdateList)
	r.GET("/download/:sole", getDownload)

	return r
}

// getUpdateList 处理获取更新列表的请求
func getUpdateList(c *gin.Context) {
	filename := c.Param("filename")
	if fileInfo, exists := fileutils.FileListJson[filename]; exists {
		c.JSON(http.StatusOK, gin.H{"ret": "ok", "appList": fileInfo})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "软件名不存在"})
	}
}

// getDownload 处理文件下载请求
func getDownload(c *gin.Context) {
	sole := c.Param("sole")

	// 从数据库获取文件路径
	dbVal, err := db.Get(sole)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sole无效"})
		return
	}

	// 构建完整文件路径
	filePath := filepath.Join("./update", dbVal)

	// 检查文件是否存在
	if !fileutil.IsExist(filePath) {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(dbVal))
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件
	c.File(filePath)
}

// ginLogger 自定义的 Gin 日志中间件
func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录错误（如果有）
		for _, e := range c.Errors {
			logger.Error("请求处理错误", zap.String("error", e.Error()))
		}

		// 计算请求处理时间
		latency := time.Since(start)

		// 记录请求信息
		logger.Info("请求日志",
			zap.String("方法", c.Request.Method),
			zap.String("路径", c.Request.URL.Path),
			zap.Int("状态码", c.Writer.Status()),
			zap.Duration("处理时间", latency),
			zap.String("客户端IP", c.ClientIP()),
		)
	}
}
