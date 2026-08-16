// Package fingerprint 把 SQL 文本归一化为用于聚类的指纹。
package fingerprint

import (
	"regexp"
	"strings"
	"unicode"
)

var opRe = regexp.MustCompile(`\s*([=<>(),])\s*`)

// Fingerprint 折叠空白、把字面量替换为占位符，并转小写。
func Fingerprint(sql string) string {
	s := strings.Join(strings.Fields(sql), " ")
	s = opRe.ReplaceAllString(s, "$1")
	s = replaceLiterals(s)
	return strings.ToLower(s)
}

func replaceLiterals(s string) string {
	runes := []rune(s)
	var b strings.Builder
	n := len(runes)
	for i := 0; i < n; i++ {
		r := runes[i]
		if r == '\'' || r == '"' {
			b.WriteString("?")
			q := r
			i++
			for i < n && runes[i] != q {
				i++
			}
			continue
		}
		if unicode.IsDigit(r) && (i == 0 || isBoundary(runes[i-1])) {
			for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				i++
			}
			b.WriteString("?")
			i-- // for 循环会自增
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isBoundary(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '(', ')', ',', '=', '<', '>', '!', '+', '-', '*', '/', ';', '.':
		return true
	}
	return false
}
