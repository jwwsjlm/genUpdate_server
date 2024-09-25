package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GetMainDirectory 获取主目录
func GetMainDirectory(path string) string {
	// 将路径转换为使用 '/' 的格式
	relativePath := filepath.ToSlash(path)

	// 分割路径并获取第一级目录
	parts := strings.Split(relativePath, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// CalculateMD5 计算文件md5
func CalculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
