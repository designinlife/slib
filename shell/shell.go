package shell

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/designinlife/slib/errors"
)

type CommandOption struct {
	Quiet         bool
	Dir           string
	Env           []string
	LineHandler   func(line string) error
	CaptureStderr bool
}

type CommandResult struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

func Run(ctx context.Context, commands []string, option *CommandOption) (*CommandResult, error) {
	cmd := exec.CommandContext(ctx, CommandName, CrossbarArg, strings.Join(commands, " && "))

	if option.Dir != "" {
		cmd.Dir = option.Dir
	}

	if len(option.Env) > 0 {
		cmd.Env = append(os.Environ(), option.Env...)
	}

	var exitError *exec.ExitError
	var exitCode int

	if option.LineHandler != nil {
		stderrPipe, err1 := cmd.StderrPipe()
		if err1 != nil {
			return &CommandResult{ExitCode: 4}, errors.Wrap(err1, "Command StderrPipe failed")
		}

		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return &CommandResult{ExitCode: 3}, errors.Wrap(err, "Command StdoutPipe failed")
		}

		if err = cmd.Start(); err != nil {
			return &CommandResult{ExitCode: 1}, errors.Wrap(err, "Command Start failed")
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				if err = option.LineHandler(strings.TrimSpace(scanner.Text())); err != nil {
					break
				}
			}
		}()
		go func() {
			defer wg.Done()

			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				if err = option.LineHandler(strings.TrimSpace(scanner.Text())); err != nil {
					break
				}
			}
		}()

		wg.Wait()

		if err = cmd.Wait(); err != nil {
			if errors.As(err, &exitError) {
				exitCode = exitError.ExitCode()
			} else if err != nil {
				exitCode = cmd.ProcessState.ExitCode()
			}

			return &CommandResult{ExitCode: exitCode}, errors.Wrapf(err, "cmd Wait failed #%d", exitCode)
		}

		return &CommandResult{}, nil
	}

	bOut := bytes.NewBuffer(nil)
	bErr := bytes.NewBuffer(nil)

	if option.Quiet {
		cmd.Stdout = bOut
		cmd.Stderr = bErr
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, bOut)
		cmd.Stderr = io.MultiWriter(os.Stderr, bErr)
	}

	// 解决 Windows 上执行 ping 命令超时不退出的问题。
	cmd.WaitDelay = 1 * time.Second

	err := cmd.Run()
	if err != nil {
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		} else if err != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}

		return &CommandResult{ExitCode: exitCode, Stderr: bErr.Bytes(), Stdout: bOut.Bytes()}, errors.Wrapf(err, "cmd Run failed #%d", exitCode)
	}

	return &CommandResult{Stderr: bErr.Bytes(), Stdout: bOut.Bytes()}, nil
}

// WaitForRun 运行命令，并等待命令执行完成。当执行成功时返回标准输出字符串。
func WaitForRun(commands []string, timeout time.Duration) (string, error) {
	if len(commands) == 0 {
		return "", fmt.Errorf("command cannot be empty")
	}

	cmdStr := strings.Join(commands, " && ")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	command := exec.CommandContext(ctx, CommandName, CrossbarArg, cmdStr)

	var stdoutBuf, stderrBuf bytes.Buffer
	command.Stdout = &stdoutBuf
	command.Stderr = &stderrBuf

	err := command.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start command '%s': %w", cmdStr, err)
	}

	err = command.Wait()

	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("command '%s' timed out after %s. Stderr: %s", cmdStr, timeout, stderrStr)
		}

		var exitErr *exec.ExitError
		ok := errors.As(err, &exitErr)
		if ok {
			return "", fmt.Errorf("command '%s' exited with error: %s. Stderr: %s", cmdStr, exitErr.Error(), stderrStr)
		}
		return "", fmt.Errorf("failed to run command '%s': %w. Stderr: %s", cmdStr, err, stderrStr)
	}

	return strings.TrimSpace(stdoutStr), nil
}
