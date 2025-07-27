package str

import "strings"

// IsTrue 检查字符串是否表示布尔值 true。
func IsTrue(s string) bool {
	switch strings.ToLower(s) {
	case "true", "t", "yes", "y", "on":
		return true
	default:
		return false
	}
}
