package fileutils

import (
	"fmt"
	ignore "github.com/Diogenesoftoronto/go-gitignore"
	json "github.com/bytedance/sonic"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/jwwsjlm/genUpdate_server/db"
	"github.com/jwwsjlm/genUpdate_server/utils"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/xid"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var FileListJson = make(map[string]FileList)
var mu sync.Mutex

func InitListUpdate(ignoreFilePath, rootDir string) (err error) {
	mu.Lock()
	defer mu.Unlock()
	//FileListJson, err = generateFileLists3(ignoreFilePath, rootDir)
	FileListJson, err = generateFileLists3(ignoreFilePath, rootDir)
	if err != nil {
		return fmt.Errorf("failed to generate file lists: %w", err)
	}

	return WriteJsonFile(FileListJson, rootDir+"/jsonBody.json")
}

// WriteJsonFile 打开jsonBody写入json文本
func WriteJsonFile(jsonBody map[string]FileList, path string) error {
	b, err := json.MarshalIndent(jsonBody, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}
	err = os.WriteFile(path, b, 0666)
	if err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}
	return nil
}

// 初始化
// generateFileLists3 生成文件列表，忽略指定文件，并组织文件信息
func generateFileLists3(ignoreFilePath, rootDir string) (map[string]FileList, error) {
	// 编译忽略文件
	ignoreFile, err := ignore.CompileIgnoreFile(ignoreFilePath)
	if err != nil {
		return nil, fmt.Errorf("编译忽略文件失败: %w", err)
	}

	// 初始化文件映射
	fileMap := make(map[string]FileList)

	// 遍历目录
	err = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("遍历路径 %v 时出错: %w", path, err)
		}

		// 忽略目录和应该被忽略的文件
		if d.IsDir() || shouldIgnoreFile(ignoreFile, path, d.Name()) {
			return nil
		}

		// 处理文件
		fileInfo, err := processFile(rootDir, path, d)
		if err != nil {
			return fmt.Errorf("处理文件 %v 时出错: %w", path, err)
		}

		// 更新文件映射
		updateFileMap(fileMap, rootDir, fileInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历路径 %v 时出错: %w", rootDir, err)
	}

	return fileMap, nil
}

// shouldIgnoreFile 判断是否应该忽略文件
func shouldIgnoreFile(ignoreFile *ignore.GitIgnore, path, name string) bool {
	return ignoreFile.MatchesPath(name) ||
		strings.HasSuffix(path, "jsonBody.json") ||
		strings.HasSuffix(path, "ReleaseNote.txt") ||
		strings.HasSuffix(path, ".ignore")
}

// processFile 处理单个文件，生成文件信息
func processFile(rootDir, path string, d os.DirEntry) (FileInfo, error) {
	// 获取文件信息
	info, err := d.Info()
	if err != nil {
		return FileInfo{}, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 计算文件的 SHA256 哈希
	sha256, err := fileutil.Sha(path, 256)
	if err != nil {
		return FileInfo{}, fmt.Errorf("计算文件哈希失败: %w", err)
	}

	// 计算相对路径
	relativePath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return FileInfo{}, fmt.Errorf("计算相对路径失败: %w", err)
	}

	// 生成唯一标识符
	guid, err := generateGUID()
	if err != nil {
		return FileInfo{}, fmt.Errorf("生成 GUID 失败: %w", err)
	}

	// 将路径信息存储到数据库，设置 TTL
	if err := db.PutWithTTL(guid, relativePath, 600); err != nil {
		return FileInfo{}, fmt.Errorf("存储文件路径失败: %w", err)
	}

	// 返回文件信息
	return FileInfo{
		Path:        relativePath,
		Name:        info.Name(),
		Size:        info.Size(),
		Sha256:      sha256,
		DownloadURL: "/download/" + guid,
	}, nil
}

// generateGUID 生成全局唯一标识符
func generateGUID() (string, error) {
	guid, err := gonanoid.New()
	if err != nil {
		// 如果 gonanoid 失败，使用 xid 作为备选
		return xid.New().String(), nil
	}
	return guid, nil
}

// updateFileMap 更新文件映射
func updateFileMap(fileMap map[string]FileList, rootDir string, fileInfo FileInfo) {
	dir := utils.GetMainDirectory(fileInfo.Path)
	if _, ok := fileMap[dir]; !ok {
		// 如果目录不存在于映射中，初始化一个新的 FileList
		fileMap[dir] = initializeFileList(rootDir, dir)
	}

	// 将文件信息添加到对应的 FileList
	fileList := fileMap[dir]
	fileList.Files = append(fileList.Files, fileInfo)
	fileMap[dir] = fileList
}

// initializeFileList 初始化文件列表，包括读取 ReleaseNote.txt
func initializeFileList(rootDir, dir string) FileList {
	dirNote := filepath.Join(rootDir, dir, "ReleaseNote.txt")

	note := ReleaseNote{
		AppName:     dir,
		Description: "null",
		Version:     "1.0.0",
	}

	// 如果存在 ReleaseNote.txt，则读取其内容
	if fileutil.IsExist(dirNote) {
		if file, err := os.ReadFile(dirNote); err == nil {
			json.Unmarshal(file, &note) // 忽略错误，因为我们有默认值
		}
	}

	return FileList{
		FileName: dir,
		Note:     note,
	}
}
