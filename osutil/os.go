package osutil

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/designinlife/slib/errors"
	"github.com/designinlife/slib/fs"
)

// IsWindows 检测当前 OS 是否为 Windows 系统。
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux 检测当前 OS 是否为 Linux 系统。
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsMachine 检查机器 ID。
func IsMachine(machineId string) bool {
	v, err := GetMachineID()
	if err != nil {
		return false
	}
	return strings.Compare(v, machineId) == 0
}

// GetMachineID 读取 Linux 系统机器 ID。
func GetMachineID() (string, error) {
	if !IsLinux() {
		return "", errors.New("GetMachineID() function only supports Linux system calls")
	}

	filename := "/etc/machine-id"

	if !fs.IsFile(filename) {
		return "", fmt.Errorf("file does not exist. (%s)", filename)
	}

	b, err := os.ReadFile(filename)
	if err != nil {
		return "", errors.Wrapf(err, "GetMachineID os ReadFile %s failed", filename)
	}

	return strings.TrimSpace(string(b)), nil
}
