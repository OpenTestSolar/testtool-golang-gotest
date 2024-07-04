package util

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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

// ParseGoVersion 解析 Go 版本号，返回主版本号和次版本号
func ParseGoVersion() (int, int, error) {
	version := runtime.Version()
	log.Println("Current Go version:", version)
	version = strings.TrimPrefix(version, "go")
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, errors.New("invalid go version format")
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return major, minor, nil
}
