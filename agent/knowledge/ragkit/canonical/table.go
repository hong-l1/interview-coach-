package canonical

import (
	"regexp"
	"strings"
)

const rulePipeTables = "normalize-pipe-tables"

var (
	pipeRowRe = regexp.MustCompile(`^\s*\|.*\|\s*$`)
	pipeSepRe = regexp.MustCompile(`^\s*\|[\s:|-]+\|\s*$`)
)

// NormalizeMarkdownPipeTables 扫描 raw markdown 中的 pipe table，
// 规范渲染每个表格并记录其 MarkdownStart/End span 到 doc.Tables。
// 返回应用的规则名列表（写入审计）。
func NormalizeMarkdownPipeTables(doc *NormalizedDocument) []string {
	raw := doc.ContentMarkdownRaw
	if raw == "" {
		raw = doc.ContentMarkdown
	}
	lines := strings.Split(raw, "\n")
	var tables []NormalizedTable
	var rendered []string
	i := 0
	byteOffset := 0
	for i < len(lines) {
		if isPipeTableStart(lines, i) {
			startByte := byteOffset
			block := collectTableLines(lines, &i, &byteOffset)
			tbl := parsePipeTable(block, startByte)
			tables = append(tables, tbl)
			rendered = append(rendered, renderPipeTable(tbl)...)
			continue
		}
		rendered = append(rendered, lines[i])
		byteOffset += len(lines[i]) + 1 // +1 for \n
		i++
	}
	doc.Tables = tables
	doc.ContentMarkdown = strings.Join(rendered, "\n")
	return []string{rulePipeTables}
}

func isPipeTableStart(lines []string, i int) bool {
	return i+1 < len(lines) && pipeRowRe.MatchString(lines[i]) && pipeSepRe.MatchString(lines[i+1])
}

func collectTableLines(lines []string, i *int, offset *int) []string {
	start := *i
	for *i < len(lines) && pipeRowRe.MatchString(lines[*i]) {
		*offset += len(lines[*i]) + 1
		*i++
	}
	return lines[start:*i]
}

func parsePipeTable(block []string, startByte int) NormalizedTable {
	rows := make([]TableRow, 0, len(block))
	for idx, line := range block {
		if idx == 1 && pipeSepRe.MatchString(line) {
			continue // 跳过分隔行
		}
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "|")
		line = strings.TrimSuffix(line, "|")
		cells := strings.Split(line, "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		rows = append(rows, TableRow(cells))
	}
	endByte := startByte
	for _, l := range block {
		endByte += len(l) + 1
	}
	return NormalizedTable{
		ID:            "",
		Page:          1,
		MarkdownStart: startByte,
		MarkdownEnd:   endByte,
		Rows:          rows,
	}
}

func renderPipeTable(tbl NormalizedTable) []string {
	if len(tbl.Rows) == 0 {
		return nil
	}
	out := make([]string, 0, len(tbl.Rows)+1)
	out = append(out, "| "+strings.Join(tbl.Rows[0], " | ")+" |")
	sep := make([]string, len(tbl.Rows[0]))
	for i := range sep {
		sep[i] = "---"
	}
	out = append(out, "| "+strings.Join(sep, " | ")+" |")
	for _, r := range tbl.Rows[1:] {
		out = append(out, "| "+strings.Join(r, " | ")+" |")
	}
	return out
}
