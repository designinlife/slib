package str

import (
	"strings"
	"unicode"
)

// IsEmpty 判断字符串是否为空。
func IsEmpty(str string) bool {
	return len(str) == 0
}

// IsBlank 判断字符串是否为空或仅包含空白字符。
func IsBlank(str string) bool {
	return strings.TrimSpace(str) == ""
}

// Trim 去除字符串两端的空白字符。
func Trim(str string) string {
	return strings.TrimSpace(str)
}

// TrimToNull 去除字符串两端的空白字符，如果结果为空则返回nil。
func TrimToNull(str string) *string {
	trimmed := strings.TrimSpace(str)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// Substring 获取子字符串。
func Substring(str string, start int) string {
	if start < 0 {
		start = 0
	}
	if start > len(str) {
		return ""
	}
	return str[start:]
}

func SubstringBetween(str, open, substr string) string {
	start := strings.Index(str, open)
	if start == -1 {
		return ""
	}
	start += len(open)
	end := strings.Index(str[start:], substr)
	if end == -1 {
		return ""
	}
	return str[start : start+end]
}

// StartsWith 判断字符串是否以指定前缀开头。
func StartsWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// EndsWith 判断字符串是否以指定后缀结尾。
func EndsWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// Reverse 反转字符串。
func Reverse(str string) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// DefaultIfEmpty 如果字符串为空，则返回默认值。
func DefaultIfEmpty(str, defaultStr string) string {
	if IsEmpty(str) {
		return defaultStr
	}
	return str
}

// Capitalize 将字符串的首字母大写。
func Capitalize(str string) string {
	if len(str) == 0 {
		return str
	}
	first := []rune(str)[0]
	return string(unicode.ToUpper(first)) + str[1:]
}

// Uncapitalize 将字符串的首字母小写。
func Uncapitalize(str string) string {
	if len(str) == 0 {
		return str
	}
	first := []rune(str)[0]
	return string(unicode.ToLower(first)) + str[1:]
}

// Abbreviate 缩写字符串。
func Abbreviate(str string, maxWidth int) string {
	if len(str) <= maxWidth {
		return str
	}
	abbrev := "..."
	if maxWidth < len(abbrev) {
		return str[:maxWidth]
	}
	return str[:maxWidth-len(abbrev)] + abbrev
}

// Difference 获取两个字符串的差异部分。
func Difference(str1, str2 string) string {
	minLength := len(str1)
	if len(str2) < minLength {
		minLength = len(str2)
	}
	for i := 0; i < minLength; i++ {
		if str1[i] != str2[i] {
			return str1[i:]
		}
	}
	if len(str1) > len(str2) {
		return str1[minLength:]
	}
	return ""
}

// IsNumeric 判断字符串是否为纯数字。
func IsNumeric(str string) bool {
	if str == "" {
		return false
	}
	for _, r := range str {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// ReverseDelimited 反转以指定分隔符分隔的字符串。
func ReverseDelimited(str, delimiter string) string {
	parts := strings.Split(str, delimiter)
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, delimiter)
}

// RightPad 在字符串右侧填充指定字符。
func RightPad(str string, size int, padChar rune) string {
	if len(str) >= size {
		return str
	}
	padding := strings.Repeat(string(padChar), size-len(str))
	return str + padding
}

// LeftPad 在字符串左侧填充指定字符。
func LeftPad(str string, size int, padChar rune) string {
	if len(str) >= size {
		return str
	}
	padding := strings.Repeat(string(padChar), size-len(str))
	return padding + str
}

// Center 将字符串居中，并用指定字符填充。
func Center(str string, size int, padChar rune) string {
	if len(str) >= size {
		return str
	}
	padding := size - len(str)
	left := padding / 2
	right := padding - left
	return strings.Repeat(string(padChar), left) + str + strings.Repeat(string(padChar), right)
}

// ContainsIgnoreCase 忽略大小写判断字符串是否包含子字符串。
func ContainsIgnoreCase(str, searchStr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(searchStr))
}

// SubstringBefore 获取子字符串之前的部分。
func SubstringBefore(str, separator string) string {
	return strings.SplitN(str, separator, 2)[0]
}

// SubstringAfter 获取子字符串之后的部分。
func SubstringAfter(str, separator string) string {
	parts := strings.SplitN(str, separator, 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// SubstringBeforeLast 获取子字符串最后一次出现之前的部分。
func SubstringBeforeLast(str, separator string) string {
	parts := strings.Split(str, separator)
	if len(parts) < 2 {
		return str
	}
	return strings.Join(parts[:len(parts)-1], separator)
}

// SubstringAfterLast 获取子字符串最后一次出现之后的部分。
func SubstringAfterLast(str, separator string) string {
	parts := strings.Split(str, separator)
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

// DeleteWhitespace 删除字符串中的所有空白字符。
func DeleteWhitespace(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(str, "\n", ""), "\r", ""), "\t", "")
}

// RemoveStartIgnoreCase 忽略大小写删除字符串开头的指定前缀。
func RemoveStartIgnoreCase(str, prefix string) string {
	if strings.HasPrefix(strings.ToLower(str), strings.ToLower(prefix)) {
		return str[len(prefix):]
	}
	return str
}

// RemoveEndIgnoreCase 忽略大小写删除字符串结尾的指定后缀。
func RemoveEndIgnoreCase(str, suffix string) string {
	if strings.HasSuffix(strings.ToLower(str), strings.ToLower(suffix)) {
		return str[:len(str)-len(suffix)]
	}
	return str
}

// Chop 删除字符串末尾的换行符或指定字符。
func Chop(str string) string {
	if len(str) == 0 {
		return str
	}
	return str[:len(str)-1]
}

// Right 获取字符串右侧的指定长度字符。
func Right(str string, size int) string {
	if len(str) <= size {
		return str
	}
	return str[len(str)-size:]
}

// Left 获取字符串左侧的指定长度字符。
func Left(str string, size int) string {
	if len(str) <= size {
		return str
	}
	return str[:size]
}

// Mid 获取字符串中间部分的字符。
func Mid(str string, pos, length int) string {
	if pos < 0 {
		pos = 0
	}
	if pos >= len(str) {
		return ""
	}
	if length < 0 {
		length = 0
	}
	end := pos + length
	if end > len(str) {
		end = len(str)
	}
	return str[pos:end]
}

// SwapCase 交换字符串中字母的大小写。
func SwapCase(str string) string {
	runes := []rune(str)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			runes[i] = unicode.ToLower(r)
		} else if unicode.IsLower(r) {
			runes[i] = unicode.ToUpper(r)
		}
	}
	return string(runes)
}

// Overlay 在指定位置覆盖字符串。
func Overlay(str, overlay string, start int, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(str) {
		end = len(str)
	}
	if start > end {
		return str
	}
	return str[:start] + overlay + str[end:]
}

// Equals 判断两个字符串是否相等。
func Equals(str1, str2 string) bool {
	return str1 == str2
}

// EqualsIgnoreCase 忽略大小写判断两个字符串是否相等。
func EqualsIgnoreCase(str1, str2 string) bool {
	return strings.EqualFold(str1, str2)
}

// DefaultString 如果字符串为空，则返回默认值。
func DefaultString(str, defaultStr string) string {
	if str == "" {
		return defaultStr
	}
	return str
}

// IsNumericSpace 判断字符串是否为纯数字（包括空格）。
func IsNumericSpace(str string) bool {
	for _, r := range str {
		if !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// IsAllUpperCase 判断字符串是否全部为大写字母。
func IsAllUpperCase(str string) bool {
	for _, r := range str {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return len(str) > 0
}

// IsAllLowerCase 判断字符串是否全部为小写字母。
func IsAllLowerCase(str string) bool {
	for _, r := range str {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return len(str) > 0
}

// IsMixedCase 判断字符串是否包含大小写混合。
func IsMixedCase(str string) bool {
	hasUpper, hasLower := false, false
	for _, r := range str {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if hasUpper && hasLower {
			return true
		}
	}
	return false
}

// ReplaceOnce 替换字符串中第一次出现的子字符串。
func ReplaceOnce(str, search, replacement string) string {
	index := strings.Index(str, search)
	if index == -1 {
		return str
	}
	return str[:index] + replacement + str[index+len(search):]
}
