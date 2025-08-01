package shell_test

import (
	"context"
	"testing"
	"time"

	"github.com/designinlife/slib/shell"
)

func TestRun(t *testing.T) {
	exitcode, stdout, stderr, err := shell.Run([]string{"ping 192.168.110.19 -n 200"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(exitcode, string(stdout), string(stderr))
}

func TestRunWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	exitcode, stdout, stderr, err := shell.RunWithContext(ctx, []string{"chcp 65001 & ping 192.168.110.19 -n 30"}, &shell.RunOption{
		Quiet: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(exitcode, string(stdout), string(stderr))
}
