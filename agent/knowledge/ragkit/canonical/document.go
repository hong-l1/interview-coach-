package canonical

// Version 标识规范化器版本，写入 CanonicalizationInfo 用于审计可比对。
const Version = "canonical-normalizer-v1"

// Source 记录文档来源元信息。
type Source struct {
	FileName   string
	FileType   string
	SourcePath string
	PageCount  int
}

// NormalizedBlock 是带 markdown 偏移的内容块，支撑 provenance 映射。
type NormalizedBlock struct {
	ID            string
	Type          string // paragraph/heading/table/list 等
	Page          int
	Text          string
	MarkdownStart int
	MarkdownEnd   int
}

// TableRow 是表格的一行单元格。
type TableRow []string

// NormalizedTable 是规范 pipe table 的 span 记录。
type NormalizedTable struct {
	ID            string
	Page          int
	MarkdownStart int
	MarkdownEnd   int
	Rows          []TableRow
}

// ParseQuality 记录解析质量信号（OCR 置信度等），本批主要留空。
type ParseQuality struct {
	Warnings   []string
	OcrLowConf bool
}

// CanonicalizationInfo 记录规范化审计信息。
type CanonicalizationInfo struct {
	Version       string
	AppliedRules []string
	RawSHA1       string
	CanonicalSHA1 string
}

// NormalizedDocument 是规范化后的中间表示，是切块的输入。
type NormalizedDocument struct {
	ContentMarkdown    string
	ContentMarkdownRaw string
	Source             Source
	Blocks             []NormalizedBlock
	Tables             []NormalizedTable
	Quality            ParseQuality
	Canonicalization   CanonicalizationInfo
}

// NewNormalizedDocument 用原始 markdown 与来源构造一个待规范化的文档。
func NewNormalizedDocument(rawMarkdown string, source Source) *NormalizedDocument {
	return &NormalizedDocument{
		ContentMarkdownRaw: rawMarkdown,
		Source:             source,
	}
}
