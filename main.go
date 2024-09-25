package main

import (
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"github.com/jwwsjlm/genUpdate_srver/auth"
	"github.com/jwwsjlm/genUpdate_srver/db"
	"github.com/jwwsjlm/genUpdate_srver/fileutils"
	"github.com/jwwsjlm/genUpdate_srver/route"
	"go.uber.org/zap"
	"os"
	"time"
)

var DownloadList = make(map[string]string)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	auth.InitLogger(zap.InfoLevel)
	defer func(Logger *zap.Logger) {
		err := Logger.Sync()
		if err != nil {
			auth.Errorf("关闭zaplogsync出错%s", err)
		}
	}(auth.Logger)
	dir, err := os.Getwd()
	if err != nil {
		auth.Panicf("获取当前目录失败:%s", err.Error())

	}
	err = db.NewRoseDb(dir + "/tmp/roseDb_basic")
	if err != nil {
		auth.Panicf("rose db数据库加载失败:%s", err.Error())
	}
	ticker := time.NewTicker(5 * time.Minute)

	defer func() {
		ticker.Stop()
		err := db.Close()
		if err != nil {
			auth.Errorf("rose 关闭失败:%s", err.Error())
		}
	}()

	dir = dir + "/update/"
	//s, _ := generateFileMap("update/.ignore", dir)
	err = fileutils.InitListUpdate("update/.ignore", dir)
	jsonData, err := json.Marshal(fileutils.FileListJson)
	auth.Infof("文件json生成:%s", jsonData)
	if err != nil {
		auth.Panicf("刷新列表失败:%s", err.Error())

	}
	go func() {
		for {
			select {
			case <-ticker.C:
				err = fileutils.InitListUpdate("update/.ignore", dir)
				if err != nil {
					auth.Errorf("刷新列表失败:%s", err.Error())

				}
			}
		}
	}()
	r := route.SetupRouter()
	err = r.Run(":8090")
	if err != nil {
		auth.Panicf("gin启动失败:%s", err.Error())

	}
}
