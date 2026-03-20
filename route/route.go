package route

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/gin-gonic/gin"
	"github.com/jwwsjlm/genUpdate_server/auth"
	"github.com/jwwsjlm/genUpdate_server/fileutils"
	"go.uber.org/zap"
)

const manifestCacheMaxAge = "60s"

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(ginLogger(auth.Logger))
	gin.SetMode(gin.ReleaseMode)

	r.GET("/updateList/:filename", getUpdateList)
	r.GET("/download/*filepath", getDownload)

	return r
}

func getUpdateList(c *gin.Context) {
	filename := c.Param("filename")
	if f, ok := fileutils.GetList(filename); ok {
		c.Header("Cache-Control", "public, max-age=60")
		c.JSON(http.StatusOK, gin.H{"ret": "ok", "appList": f})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"ret": "error", "error": "软件名不存在"})
}

func getDownload(c *gin.Context) {
	relPath := strings.TrimPrefix(c.Param("filepath"), "/")
	if relPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ret": "error", "error": "文件路径不能为空"})
		return
	}
	if strings.Contains(relPath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"ret": "error", "error": "非法文件路径"})
		return
	}

	cleanPath := filepath.Clean(relPath)
	filePath := filepath.Join("./update", cleanPath)
	if !fileutil.IsExist(filePath) {
		c.JSON(http.StatusNotFound, gin.H{"ret": "error", "error": "文件不存在"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(cleanPath))
	c.Header("Content-Type", "application/octet-stream")
	c.File(filePath)
}

func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		for _, e := range c.Errors {
			logger.Error("请求处理错误", zap.String("error", e.Error()))
		}

		latency := time.Since(start)
		logger.Info("请求日志",
			zap.String("方法", c.Request.Method),
			zap.String("路径", c.Request.URL.Path),
			zap.Int("状态码", c.Writer.Status()),
			zap.Duration("处理时间", latency),
			zap.String("客户端IP", c.ClientIP()),
		)
	}
}
