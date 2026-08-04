package canonical

import (
	"crypto/sha1"
	"encoding/hex"
)

// Normalize 在 doc 上运行完整规则链，原地填充 ContentMarkdown、Tables、Canonicalization。
// 返回 AppliedRules 列表。
func Normalize(doc *NormalizedDocument) []string {
	raw := doc.ContentMarkdownRaw
	if raw == "" {
		raw = doc.ContentMarkdown
	}
	rawSHA := sha1Hex(raw)

	applied := []string{"normalize-unicode-lineendings"}
	s := normalizeUnicodeAndLineEndings(raw)
	s = normalizeLineNoise(s)
	applied = append(applied, "normalize-line-noise")
	s = normalizeHeadings(s)
	applied = append(applied, "normalize-headings")
	s = normalizeHeadingContinuations(s)
	applied = append(applied, "normalize-heading-continuations")
	s = collapseBlankLines(s)
	applied = append(applied, "collapse-blank-lines")

	doc.ContentMarkdown = s
	doc.ContentMarkdownRaw = "" // force NormalizeMarkdownPipeTables to use already-normalized content
	tblRules := NormalizeMarkdownPipeTables(doc)
	applied = append(applied, tblRules...)

	canonicalSHA := sha1Hex(doc.ContentMarkdown)
	doc.Canonicalization = CanonicalizationInfo{
		Version:       Version,
		AppliedRules:  applied,
		RawSHA1:       rawSHA,
		CanonicalSHA1: canonicalSHA,
	}
	return applied
}

func sha1Hex(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
