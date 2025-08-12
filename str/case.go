package str

import (
	"strings"
	"unicode"
)

// ToPascalCase 转换为帕斯卡命名形式。
func ToPascalCase(s string) string {
	return ToCamelCase(s, true)
}

// ToCamelCase 转换为驼峰命名形式。
func ToCamelCase(s string, firstCharUpper bool) string {
	if s == "" {
		return s
	}

	// 分隔符：下划线、空格、连字符等
	parts := splitToWords(s)

	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 && !firstCharUpper {
			parts[i] = strings.ToLower(p[:1]) + p[1:]
		} else {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, "")
}

// 按分隔符拆分成单词。
func splitToWords(s string) []string {
	var words []string
	var current []rune

	for _, r := range s {
		if r == '_' || r == '-' || unicode.IsSpace(r) {
			if len(current) > 0 {
				words = append(words, string(current))
				current = current[:0]
			}
			continue
		}

		if unicode.IsUpper(r) && len(current) > 0 && !unicode.IsUpper(rune(current[len(current)-1])) {
			// 遇到大写，视为新单词
			words = append(words, string(current))
			current = current[:0]
		}

		current = append(current, r)
	}

	if len(current) > 0 {
		words = append(words, string(current))
	}

	return words
}
