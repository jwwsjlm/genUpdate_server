package db

import (
	"time"

	"github.com/rosedblabs/rosedb/v2"
)

// sdb 是 RoseDB 实例的全局变量
var sdb *rosedb.DB

// NewRoseDb 初始化并打开一个新的 RoseDB 实例
func NewRoseDb(dirPath string) error {
	options := rosedb.DefaultOptions
	options.DirPath = dirPath

	var err error
	sdb, err = rosedb.Open(options)
	if err != nil {
		return err
	}
	return nil
}

// PutWithTTL 存储带有 TTL 的键值对
func PutWithTTL(key, value string, ttlSeconds int) error {
	return sdb.PutWithTTL([]byte(key), []byte(value), time.Duration(ttlSeconds)*time.Second)
}

// Get 根据键获取值
func Get(key string) (string, error) {
	value, err := sdb.Get([]byte(key))
	return string(value), err
}

// Exist 检查键是否存在
func Exist(key string) (bool, error) {
	return sdb.Exist([]byte(key))
}

// Close 同步并关闭数据库
func Close() error {
	if err := sdb.Sync(); err != nil {
		return err
	}
	return sdb.Close()
}
