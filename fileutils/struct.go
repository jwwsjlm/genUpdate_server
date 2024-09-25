package fileutils

type FileInfo struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Sha256      string `json:"sha256"`
	DownloadURL string `json:"downloadURL"`
}
type FileList struct {
	FileName string      `json:"fileName"`
	Note     ReleaseNote `json:"ReleaseNote"`
	Files    []FileInfo  `json:"fileList"`
}
type ReleaseNote struct {
	AppName     string `json:"appName"`
	Description string `json:"description"`
	Version     string `json:"version"`
}
