package str

import (
	"os"
	"regexp"

	"github.com/mattn/go-isatty"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func StripAnsi(str string) string {
	return re.ReplaceAllString(str, "")
}

func AutoStripAnsi(str string) string {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		return str
	}
	return re.ReplaceAllString(str, "")
}
