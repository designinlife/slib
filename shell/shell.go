package shell

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type RunOption struct {
	Quiet         bool
	Env           []string
	LineHandler   func(line string) error
	CaptureStderr bool
}

func defaultRunOption() *RunOption {
	return &RunOption{
		Quiet: true,
	}
}

func Run(commands []string) (int, []byte, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return RunWithContext(ctx, commands, defaultRunOption())
}

func RunWithContext(ctx context.Context, commands []string, option *RunOption) (int, []byte, []byte, error) {
	cmd := exec.CommandContext(ctx, CommandName, CrossbarArg, strings.Join(commands, " && "))

	if len(option.Env) > 0 {
		cmd.Env = append(os.Environ(), option.Env...)
	}

	var exitError *exec.ExitError
	var exitCode int

	if option.LineHandler != nil {
		stderrPipe, err1 := cmd.StderrPipe()
		if err1 != nil {
			return 4, nil, nil, err1
		}

		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return 3, nil, nil, err
		}

		if err = cmd.Start(); err != nil {
			return 1, nil, nil, err
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

			return exitCode, nil, nil, err
		}

		return 0, nil, nil, nil
	} else {
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

			return exitCode, bOut.Bytes(), bErr.Bytes(), fmt.Errorf("command: %s, exit: %d: %w", strings.Join(commands, " && "), exitCode, err)
		}

		return 0, bOut.Bytes(), bErr.Bytes(), nil
	}
}
