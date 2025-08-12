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

// SearchFile 在若干目录中搜索 name 文件。
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
