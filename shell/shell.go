package shell

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/designinlife/slib/errors"
)

type RunOption struct {
	Quiet         bool
	Env           []string
	LineHandler   func(line string) error
	CaptureStderr bool
}

type CommandResult struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

func defaultRunOption() *RunOption {
	return &RunOption{
		Quiet: true,
	}
}

func Run(commands []string) (*CommandResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return RunWithContext(ctx, commands, defaultRunOption())
}

func RunWithContext(ctx context.Context, commands []string, option *RunOption) (*CommandResult, error) {
	cmd := exec.CommandContext(ctx, CommandName, CrossbarArg, strings.Join(commands, " && "))

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
