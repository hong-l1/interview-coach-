package canonical

import "testing"

func TestNormalizeUnicodeAndLineEndings(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a\r\nb", "a\nb"},
		{"a\rb", "a\nb"},
		{"ａ\r\n", "a\n"}, // NFKC: U+FF41 fullwidth a → U+0061 regular a
	}
	for _, c := range cases {
		if got := normalizeUnicodeAndLineEndings(c.in); got != c.want {
			t.Errorf("normalizeUnicodeAndLineEndings(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeLineNoise(t *testing.T) {
	cases := []struct{ in, want string }{
		{"## · title\n", "## title\n"},
		{"- · item\n", "- item\n"},
		{"- - item\n", "- item\n"},
		{"normal text", "normal text"},
	}
	for _, c := range cases {
		if got := normalizeLineNoise(c.in); got != c.want {
			t.Errorf("normalizeLineNoise(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeHeadings(t *testing.T) {
	cases := []struct{ in, want string }{
		{"#title\n", "# title\n"},            // ATX 补空格
		{"## **bold**\n", "## bold\n"},        // 去强调标记
		{"## 1.1标题\n", "## 1.1 标题\n"},      // 数字前缀补空格
		{"## 标　题\n", "## 标题\n"},            // CJK 全角空格去掉
		{"## 标 题\n", "## 标题\n"},            // CJK 间空格去掉
	}
	for _, c := range cases {
		if got := normalizeHeadings(c.in); got != c.want {
			t.Errorf("normalizeHeadings(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeHeadingContinuations(t *testing.T) {
	// "1.1\nTERM" 跨行碎片合并为 "## 1.1 TERM"
	in := "# Chapter\n\n## 1.1\nTERM\nbody"
	want := "# Chapter\n\n## 1.1 TERM\nbody"
	if got := normalizeHeadingContinuations(in); got != want {
		t.Errorf("got=%q want=%q", got, want)
	}
}

func TestCollapseBlankLines(t *testing.T) {
	in := "a\n\n\n\n\nb"
	want := "a\n\nb"
	if got := collapseBlankLines(in); got != want {
		t.Errorf("got=%q want=%q", got, want)
	}
}
