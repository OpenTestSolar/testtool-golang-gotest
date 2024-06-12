package util

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func ElementIsInSlice(element string, elements []string) bool {
	for _, item := range elements {
		if element == item {
			return true
		}
	}
	return false
}

// 根据给定路径返回路径对应的目录以及文件名，若路径指向目录则仅返回目录，若路径指向文件则返回文件对应目录以及文件名，如果路径不存在则返回错误
func GetPathAndFileName(projPath, path string) (dir string, file string, err error) {
	if path == "" {
		return "", "", errors.New("path is empty")
	}
	absPath := filepath.Join(projPath, path)
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return "", "", err
	}
	if fileInfo.IsDir() {
		return path, "", nil
	} else {
		relPath, err := filepath.Rel(projPath, filepath.Dir(absPath))
		if err != nil {
			return "", "", err
		}
		return relPath, filepath.Base(absPath), nil
	}
}

func GetWorkspace(path string) string {
	var projPath string
	if path != "" {
		projPath = path
	} else {
		projPath = os.Getenv("TESTSOLAR_WORKSPACE")
	}
	return strings.TrimSuffix(projPath, string(os.PathSeparator))
}
