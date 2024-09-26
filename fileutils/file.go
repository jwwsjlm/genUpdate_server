package fileutils

import (
	"fmt"
	ignore "github.com/Diogenesoftoronto/go-gitignore"
	json "github.com/bytedance/sonic"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/jwwsjlm/genUpdate_srver/auth"
	"github.com/jwwsjlm/genUpdate_srver/db"
	"github.com/jwwsjlm/genUpdate_srver/utils"
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
		return err
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
func generateFileLists3(ignoreFilePath, rootDir string) (map[string]FileList, error) {
	ignoreFile, err := ignore.CompileIgnoreFile(ignoreFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to compile ignore file: %w", err)
	}
	fileMap := make(map[string]FileList)
	err = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {

		if err != nil {
			return err
		}
		//目录的话跳过
		if d.IsDir() {
			return nil
		}
		//目录和忽略列表当中的文件不进行操作
		if ignoreFile.MatchesPath(path) {
			return nil
		}
		//忽略列表不进行操作
		if strings.HasSuffix(ignoreFilePath, d.Name()) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		sha256, err := fileutil.Sha(path, 256)
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		guid, err := gonanoid.New()
		if err != nil {
			guid = xid.New().String()
		}
		//guid := xid.New().String()
		err = db.PutWithTTL(guid, relativePath, 600)

		if err != nil {
			return err
		}

		fileInfo := FileInfo{
			Path:        relativePath,
			Name:        info.Name(),
			Size:        info.Size(),
			Sha256:      sha256,
			DownloadURL: "/download/" + guid,
		}

		dir := utils.GetMainDirectory(relativePath)
		// 如果目录不存在于 map 中，初始化一个新的 代表为新的软件
		if _, ok := fileMap[dir]; !ok {
			dirNote := rootDir + dir + "/ReleaseNote.txt"
			auth.Infof(dirNote)
			Note := ReleaseNote{}
			if fileutil.IsExist(dirNote) {
				//找到ReleaseNote.txt文件,使用自定义配置
				file, err := os.ReadFile(dirNote)
				if err != nil {
					return err
				}
				err = json.Unmarshal(file, &Note)
				if err != nil {
					return err
				}

			} else {
				//未找到ReleaseNote.txt文件,使用默认配置
				Note.AppName = dir
				Note.Description = "null"
				Note.Version = "1.0.0"
			}
			fileMap[dir] = FileList{
				FileName: dir,
				Note:     Note,
			}
		}

		// 将 FileInfo 添加到对应的 FileList
		fileList := fileMap[dir]
		fileList.Files = append(fileList.Files, fileInfo)
		fileMap[dir] = fileList

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the path %v: %w", rootDir, err)
	}

	return fileMap, nil
}
