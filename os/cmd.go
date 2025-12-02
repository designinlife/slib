package os

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
)

// CommandOption 定义命令执行的选项
type CommandOption struct {
	Env     []string  // 额外的环境变量，格式 "KEY=VALUE"
	Silent  bool      // 是否静默执行（不输出日志）
	Printer io.Writer // 日志输出的位置 (例如 os.Stdout)
	Dir     string    // 指定命令执行的工作目录
}

// CommandResult 定义命令执行的结果
type CommandResult struct {
	ExitCode int    // 退出码，0 表示成功
	Stdout   []byte // 标准输出内容
	Stderr   []byte // 错误输出内容
}

// Run 执行 Shell 命令，支持 Windows 和 Linux，支持 string 或 []string
func Run[C string | []string](ctx context.Context, option *CommandOption, commands C) (*CommandResult, error) {
	if option == nil {
		option = &CommandOption{}
	}

	var shellName string
	var shellFlag string

	if runtime.GOOS == "windows" {
		shellName = "cmd"
		shellFlag = "/C"
	} else {
		shellName = "bash"
		shellFlag = "-c"
	}

	var commandStr string
	switch v := any(commands).(type) {
	case string:
		commandStr = v
	case []string:
		commandStr = strings.Join(v, " && ")
	default:
		return &CommandResult{ExitCode: -1}, fmt.Errorf("unsupported command type")
	}

	cmd := exec.CommandContext(ctx, shellName, shellFlag, commandStr)

	if option.Dir != "" {
		cmd.Dir = option.Dir
	}

	if len(option.Env) > 0 {
		cmd.Env = slices.Concat(option.Env, os.Environ())
	} else {
		cmd.Env = os.Environ()
	}

	var stdout, stderr bytes.Buffer

	if option.Silent {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	} else {
		if option.Printer != nil {
			cmd.Stdout = io.MultiWriter(option.Printer, &stdout)
			cmd.Stderr = io.MultiWriter(option.Printer, &stderr)
		} else {
			cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
			cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
		}
	}

	err := cmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return &CommandResult{ExitCode: exitError.ExitCode(), Stderr: stderr.Bytes(), Stdout: stdout.Bytes()}, err
		}
		// 其他错误（如路径找不到，Context取消等）
		return &CommandResult{ExitCode: 1, Stderr: stderr.Bytes(), Stdout: stdout.Bytes()}, fmt.Errorf("failed to run command: %w", err)
	}

	return &CommandResult{ExitCode: 0, Stderr: stderr.Bytes(), Stdout: stdout.Bytes()}, nil
}
