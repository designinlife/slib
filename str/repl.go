package str

import "strings"

// ReplaceAllFullWidthChars 替换全角字符为半角。
func ReplaceAllFullWidthChars(s string) string {
	return strings.NewReplacer("，", ",", "（", "(", "）", ")", "　", " ").Replace(s)
}
