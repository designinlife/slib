package fs

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/designinlife/slib/errors"

	"github.com/mitchellh/go-homedir"
)

// IsFile 检查是否文件？
func IsFile(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// IsDir 检查是否文件夹？
func IsDir(dirname string) bool {
	info, err := os.Stat(dirname)
	if err != nil {
		return false
	}

	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}

// IsDirEmpty An empty directory means it contains no regular files, hidden files, or subdirectories.
// It returns true if the directory is empty, false otherwise, and an error if there's a problem accessing the directory.
func IsDirEmpty(dirPath string) (bool, error) {
	// Check if the path exists and is a directory
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("directory '%s' does not exist", dirPath)
		}
		return false, fmt.Errorf("failed to get directory info for '%s': %w", dirPath, err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("path '%s' is not a directory", dirPath)
	}

	// Use os.ReadDir to get directory entries.
	// os.ReadDir returns a slice of fs.DirEntry, which allows checking type
	// without stat'ing each entry individually.
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, fmt.Errorf("failed to read directory '%s': %w", dirPath, err)
	}

	// If the slice is empty, the directory is empty.
	return len(entries) == 0, nil
}

// SearchFile 在 dirs 目录列表中搜索 name 文件。
func SearchFile(name string, dirs []string) (string, error) {
	for _, v := range dirs {
		fn, err := homedir.Expand(path.Join(v, name))
		if err != nil {
			return "", errors.Wrap(err, "SearchFile homedir Expand failed")
		}

		if IsFile(fn) {
			return fn, nil
		}
	}

	return "", fmt.Errorf("no file was found. (%s in %s)", name, strings.Join(dirs, ", "))
}

// FileSize 获取文件大小。
func FileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, errors.Wrapf(err, "FileSize %s failed", filePath)
	}

	return fileInfo.Size(), nil
}
