package route

import (
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/gin-gonic/gin"
	"github.com/jwwsjlm/genUpdate_srver/auth"
	"github.com/jwwsjlm/genUpdate_srver/db"
	"github.com/jwwsjlm/genUpdate_srver/fileutils"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
	"time"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(ginLogger(auth.Logger))
	gin.SetMode(gin.ReleaseMode)
	r.GET("updateList/:filename", getUpdateList)
	r.GET("/download/:sole", getDownload)
	return r
}
func getUpdateList(c *gin.Context) {
	filename := c.Param("filename")
	if fileInfo, exists := fileutils.FileListJson[filename]; exists {
		c.JSON(http.StatusOK, gin.H{"ret": "ok", "appList": fileInfo})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "软件名不存在"})
	}
}
func getDownload(c *gin.Context) {
	filename := c.Param("sole")
	var err error
	var dbVal string
	if dbVal, err = db.Get(filename); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sole无效"})
		c.Abort()
	}
	//获取文件路径
	filePath := filepath.Join("./update", dbVal)
	var file string
	file = filePath
	//if file, err = filepath.Abs(filePath); err != nil {
	//	c.JSON(http.StatusNotFound, gin.H{"error": "路径不存在"})
	//	c.Abort()
	//}

	if ok := fileutil.IsExist(file); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		c.Abort()
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(dbVal))
	c.Header("Content-Type", "application/octet-stream")
	c.File(file)

}
func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("Error occurred", zap.String("error", e.Error()))
			}
		}
		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)

	}
}
