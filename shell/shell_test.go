package shell_test

import (
	"context"
	"testing"
	"time"

	"github.com/designinlife/slib/shell"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	cr, err := shell.Run(ctx, []string{"chcp 65001 & ping 192.168.110.19 -n 30"}, &shell.CommandOption{
		Quiet: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cr.ExitCode, string(cr.Stdout), string(cr.Stderr))
}
