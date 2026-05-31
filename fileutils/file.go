package fileutils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	ignore "github.com/Diogenesoftoronto/go-gitignore"
	"github.com/jwwsjlm/genUpdate_server/utils"
)

var (
	listJSON      = make(map[string]FileList)
	filePathIndex = make(map[string]struct{})
	mu            sync.RWMutex
)

type cachedFileMeta struct {
	Size    int64  `json:"size"`
	ModTime int64  `json:"modTime"`
	SHA256  string `json:"sha256"`
}

type manifestCache struct {
	Files map[string]cachedFileMeta `json:"files"`
}

func GetList(fn string) (FileList, bool) {
	mu.RLock()
	defer mu.RUnlock()
	fileInfo, ok := listJSON[fn]
	return fileInfo, ok
}

func GetAllLists() []FileList {
	mu.RLock()
	defer mu.RUnlock()

	lists := make([]FileList, 0, len(listJSON))
	for _, fileInfo := range listJSON {
		lists = append(lists, fileInfo)
	}
	sort.Slice(lists, func(i, j int) bool {
		return lists[i].FileName < lists[j].FileName
	})
	return lists
}

func HasFilePath(path string) bool {
	cleanPath := filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimPrefix(path, "/"))))
	if cleanPath == "." || cleanPath == "" {
		return false
	}

	mu.RLock()
	defer mu.RUnlock()
	_, ok := filePathIndex[cleanPath]
	return ok
}

func GetJSONText() (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	jsonData, err := json.Marshal(listJSON)
	return string(jsonData), err
}

func InitListUpdate(ignoreFilePath, rootDir string) error {
	cachePath := filepath.Join(rootDir, "manifest-cache.json")
	cache, err := loadManifestCache(cachePath)
	if err != nil {
		return fmt.Errorf("failed to load manifest cache: %w", err)
	}

	newList, newCache, err := generateFileLists(ignoreFilePath, rootDir, cache)
	if err != nil {
		return fmt.Errorf("failed to generate file lists: %w", err)
	}

	if err := saveManifestCache(cachePath, newCache); err != nil {
		return fmt.Errorf("failed to save manifest cache: %w", err)
	}
	if err := writeJSONFile(filepath.Join(rootDir, "jsonBody.json"), newList); err != nil {
		return err
	}

	mu.Lock()
	listJSON = newList
	filePathIndex = buildFilePathIndex(newList)
	mu.Unlock()
	return nil
}

func buildFilePathIndex(lists map[string]FileList) map[string]struct{} {
	index := make(map[string]struct{})
	for _, fileInfo := range lists {
		for _, f := range fileInfo.Files {
			if f.Path != "" {
				index[f.Path] = struct{}{}
			}
		}
	}
	return index
}

func writeJSONFile(path string, list map[string]FileList) error {
	b, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}
	if err := os.WriteFile(path, b, 0o666); err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}
	return nil
}

func loadManifestCache(path string) (manifestCache, error) {
	cache := manifestCache{Files: make(map[string]cachedFileMeta)}
	if !fileExists(path) {
		return cache, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return cache, err
	}
	if len(b) == 0 {
		return cache, nil
	}
	if err := json.Unmarshal(b, &cache); err != nil {
		return cache, err
	}
	if cache.Files == nil {
		cache.Files = make(map[string]cachedFileMeta)
	}
	return cache, nil
}

func saveManifestCache(path string, cache manifestCache) error {
	b, err := json.MarshalIndent(cache, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o666)
}

func generateFileLists(ignoreFilePath, rootDir string, cache manifestCache) (map[string]FileList, manifestCache, error) {
	ignoreFile, err := compileIgnoreFile(ignoreFilePath)
	if err != nil {
		return nil, manifestCache{}, fmt.Errorf("编译忽略文件失败: %w", err)
	}

	fileMap := make(map[string]FileList)
	newCache := manifestCache{Files: make(map[string]cachedFileMeta)}

	err = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("遍历路径 %v 时出错: %w", path, walkErr)
		}
		if path == rootDir {
			return nil
		}
		if shouldIgnorePath(ignoreFile, rootDir, path, d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}

		fileInfo, cacheMeta, err := processFile(rootDir, path, d, cache)
		if err != nil {
			return fmt.Errorf("处理文件 %v 时出错: %w", path, err)
		}

		newCache.Files[fileInfo.Path] = cacheMeta
		updateFileMap(fileMap, rootDir, fileInfo)
		return nil
	})
	if err != nil {
		return nil, manifestCache{}, fmt.Errorf("遍历路径 %v 时出错: %w", rootDir, err)
	}

	return fileMap, newCache, nil
}

func compileIgnoreFile(ignoreFilePath string) (*ignore.GitIgnore, error) {
	if !fileExists(ignoreFilePath) {
		return ignore.CompileIgnoreLines(), nil
	}
	return ignore.CompileIgnoreFile(ignoreFilePath)
}

func shouldIgnorePath(ignoreFile *ignore.GitIgnore, rootDir, path, name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "jsonBody.json", "manifest-cache.json", "ReleaseNote.txt":
		return true
	}

	relativePath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return true
	}
	relativePath = filepath.ToSlash(relativePath)
	return ignoreFile.MatchesPath(name) || ignoreFile.MatchesPath(relativePath)
}

func processFile(rootDir, path string, d os.DirEntry, cache manifestCache) (FileInfo, cachedFileMeta, error) {
	info, err := d.Info()
	if err != nil {
		return FileInfo{}, cachedFileMeta{}, fmt.Errorf("获取文件信息失败: %w", err)
	}

	relativePath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return FileInfo{}, cachedFileMeta{}, fmt.Errorf("计算相对路径失败: %w", err)
	}
	relativePath = filepath.ToSlash(relativePath)

	meta := cachedFileMeta{
		Size:    info.Size(),
		ModTime: info.ModTime().UnixNano(),
	}

	if old, ok := cache.Files[relativePath]; ok && old.Size == meta.Size && old.ModTime == meta.ModTime {
		meta.SHA256 = old.SHA256
	} else {
		meta.SHA256, err = calculateSHA256(path)
		if err != nil {
			return FileInfo{}, cachedFileMeta{}, fmt.Errorf("计算文件哈希失败: %w", err)
		}
	}

	return FileInfo{
		Path:        relativePath,
		Name:        info.Name(),
		Size:        info.Size(),
		Sha256:      meta.SHA256,
		DownloadURL: "/download/" + relativePath,
		ModTime:     info.ModTime().UTC().Format(time.RFC3339),
	}, meta, nil
}

func updateFileMap(fileMap map[string]FileList, rootDir string, fileInfo FileInfo) {
	dir := utils.GetMainDirectory(fileInfo.Path)
	if _, ok := fileMap[dir]; !ok {
		fileMap[dir] = initializeFileList(rootDir, dir)
	}

	fileList := fileMap[dir]
	fileList.Files = append(fileList.Files, fileInfo)
	fileMap[dir] = fileList
}

func initializeFileList(rootDir, dir string) FileList {
	dirNote := filepath.Join(rootDir, dir, "ReleaseNote.txt")
	note := ReleaseNote{
		AppName:     dir,
		Description: "null",
		Version:     "1.0.0",
	}

	if fileExists(dirNote) {
		if file, err := os.ReadFile(dirNote); err == nil {
			_ = json.Unmarshal(file, &note)
		}
	}

	return FileList{
		FileName: dir,
		Note:     note,
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func calculateSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
