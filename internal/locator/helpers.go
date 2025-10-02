package locator

import (
	"fmt"
	"os"
)

// findNewestFile 返回文件列表中修改时间最新的文件
func findNewestFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("file list is empty")
	}

	var newestFile string
	var newestTime int64

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		modTime := info.ModTime().Unix()
		if modTime > newestTime {
			newestTime = modTime
			newestFile = file
		}
	}

	if newestFile == "" {
		return "", fmt.Errorf("no accessible files found")
	}

	return newestFile, nil
}
