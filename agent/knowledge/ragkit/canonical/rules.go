package canonical

import (
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

func normalizeUnicodeAndLineEndings(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return norm.NFKC.String(s)
}

var noisePatterns = []struct{ pat, repl string }{
	{"## · ", "## "},
	{"- · ", "- "},
	{"- - ", "- "},
}

func normalizeLineNoise(s string) string {
	out := s
	for _, p := range noisePatterns {
		out = strings.ReplaceAll(out, p.pat, p.repl)
	}
	return out
}

var (
	atxHeadingRe = regexp.MustCompile(`(?m)^(#{1,6})([^\s#].*)$`)
	emphasisBoldRe = regexp.MustCompile(`(?m)^(\s*#{1,6}\s*)\*\*(.+?)\*\*[^\S\n]*$`)
	emphasisUlRe   = regexp.MustCompile(`(?m)^(\s*#{1,6}\s*)__(.+?)__\s*$`)
	cjkSpaceRe     = regexp.MustCompile(`(?m)^(\s*#{1,6}\s*)(.*)$`)
)

func normalizeHeadings(s string) string {
	// 1. ATX 补空格：#title -> # title
	s = atxHeadingRe.ReplaceAllStringFunc(s, func(m string) string {
		loc := atxHeadingRe.FindStringSubmatchIndex(m)
		hashes := m[loc[2]:loc[3]]
		rest := m[loc[4]:loc[5]]
		return hashes + " " + rest
	})
	// 2. 去强调标记
	s = emphasisBoldRe.ReplaceAllString(s, "$1$2")
	s = emphasisUlRe.ReplaceAllString(s, "$1$2")
	// 3. CJK 内部去空格（标题行）+ 数字前缀补空格
	s = cjkSpaceRe.ReplaceAllStringFunc(s, func(m string) string {
		loc := cjkSpaceRe.FindStringSubmatchIndex(m)
		prefix := m[loc[2]:loc[3]]
		rest := m[loc[4]:loc[5]]
		rest = removeCJKSpaces(rest)
		rest = addSpaceAfterNumPrefix(rest)
		return prefix + rest
	})
	return s
}

func removeCJKSpaces(s string) string {
	// Remove fullwidth spaces (U+3000) and regular spaces between CJK characters.
	return regexp.MustCompile(`([\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}])[\s　]+([\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}])`).ReplaceAllString(s, "$1$2")
}

func addSpaceAfterNumPrefix(s string) string {
	return regexp.MustCompile(`^(\d+(?:\.\d+)+)([^\s\d])`).ReplaceAllString(s, "$1 $2")
}

var headingContRe = regexp.MustCompile(`(?m)^(#{1,6}\s*)([0-9]+(?:\.[0-9]+)*)\s*\n([^\n#].*)$`)

func normalizeHeadingContinuations(s string) string {
	return headingContRe.ReplaceAllString(s, "$1$2 $3")
}

var multiBlankRe = regexp.MustCompile(`\n{3,}`)

func collapseBlankLines(s string) string {
	return multiBlankRe.ReplaceAllString(s, "\n\n")
}
