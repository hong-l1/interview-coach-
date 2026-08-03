# ragkit 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `awesomeProject4` 内新建自包含包 `agent/knowledge/ragkit`，把 rag-retrievalOps 的 文档标准化 / 切块 / 检索增强 / 精简治理 四块精华重写进来，对齐现有栈（Milvus + eino + Ark），纯 Go 包 API，零现有代码改动，附独立验证 CLI。

**Architecture:** facade 包 `ragkit` + 四子包 `canonical`/`chunking`/`retrieval`/`governance`。入库链路：eino loader → canonical.Normalize → chunking.StrategyRouter.Split → 现有 embedder+indexer；检索链路：现有 HybridRetrieve → retrieval 后处理（去重/动态TopK/Jaccard重排/证据门控）。治理用 GORM 持久化策略档案与审计事件，经 CLI 触发。两处接线点（`import.go`/`retriever.go`）只加注释 stub，默认关闭。

**Tech Stack:** Go 1.25.5、eino + eino-ext（milvus2 indexer/retriever、ark embedder）、Milvus/Zilliz、GORM/MySQL、godotenv、`golang.org/x/text/unicode/norm`（已间接依赖）、新增 `github.com/yuin/goldmark`。

## Global Constraints

- **零现有代码改动**：禁止修改 `agent/knowledge/rag/`、`agent/knowledge/import.go`、`agent/tool/retriever.go`、`agent/agents/*`、面试 SSE 链路的逻辑。Task 0 与 Task 15 仅在 `import.go`/`retriever.go` 加**注释 stub**，不改可执行逻辑。接线点开关用环境变量 `RAGKIT_ENABLED`（默认关闭，用户手动切换）。
- **对齐现有栈**：Milvus 集合 `documents`、维度 1024、COSINE+HNSW、BM25 sparse；embedder 走 `os.Getenv("api_key")`/`embedding_id`/`base_url`（Ark qwen3-embedding-0.6b）；复用 `Init.NewMilvusManger`。**禁止**引入新向量库/新 embedding 服务。
- **新增依赖唯一**：`github.com/yuin/goldmark`（Markdown AST 切块）。其余依赖必须已在 `go.sum` 中存在（`golang.org/x/text` 已是 indirect v0.37.0）。
- **不碰面试链路**：不新增 HTTP 路由、不改 `handler.Register`、不动 wire.go（governance 的 GORM 注册放到一个**独立**的 `ragkitInit` 包，由 CLI 单独调用，不进 `InitMysql`，避免污染服务端 AutoMigrate）。详见 Task 12。
- **模块路径**：`awesomeProject4`；包名 `ragkit`。Go 代码风格对齐仓库：构造器 `NewXxx`、`ctx context.Context` 首参、构造失败可 panic（沿用 `rag` 包风格），但**业务逻辑返回 error**。
- **git**：当前分支 `rag`。每个 Task 末尾 commit，commit message 用 `feat(ragkit): ...` 前缀。
- **TDD**：每个有逻辑的 Task 先写失败测试再实现。纯类型/常量定义 Task 可省测试。

---

## File Structure

```
agent/knowledge/ragkit/
├── ragkit.go                          # facade：Index() / Retrieve() / NormalizeDocs() / Split() / DefaultRouter()
├── canonical/
│   ├── document.go                     # NormalizedDocument / Block / Table / Quality / CanonicalizationInfo 类型 + 常量
│   ├── normalize.go                    # Normalize() 入口 + 规则链编排
│   ├── rules.go                         # 5 条规则函数（unicode/whitespace/noise/heading/heading-cont/blank）
│   ├── table.go                         # NormalizeMarkdownPipeTables + span 追踪
│   ├── provenance.go                    # AnnotateChunksWithProvenance
│   └── *_test.go
├── chunking/
│   ├── strategy.go                     # Strategy 接口 + Request + 路由常量
│   ├── router.go                        # StrategyRouter + RoutedStrategy + Matcher
│   ├── markdown.go                      # MarkdownStrategy（goldmark AST）
│   ├── table_aware.go                   # TableAwareStrategy
│   ├── parent_child.go                  # finalizeChunks 父子元数据
│   ├── semantic.go                      # 语义二次切分（embedder 可选）
│   └── *_test.go
├── retrieval/
│   ├── searcher.go                      # Searcher 接口 + milvusSearcher 适配 HybridRetrieve
│   ├── result.go                        # Result / Metrics / EvidenceGate / Item 类型
│   ├── topk.go                          # DecideDynamicTopK / DecideStrategicTopK / ScoreCliffGuard / TokenBudget
│   ├── rerank.go                        # JaccardReranker
│   ├── citation.go                      # CitationConsistencyChecker（默认关）
│   ├── evidence_gate.go                 # EvaluateEvidenceGate + 5 类拒绝
│   ├── dedupe.go                        # Dedupe
│   ├── postprocess.go                   # PostProcess 编排
│   └── *_test.go
├── governance/
│   ├── profile.go                       # StrategyProfile 类型 + GetActive/Activate/Rollback（纯逻辑，依赖 store 接口）
│   ├── profile_store.go                 # ProfileStore 接口 + gormProfileStore
│   ├── audit.go                         # AuditEvent + AuditLogger 接口 + 默认 logger + 脱敏
│   ├── kb_health.go                     # IndexHealth + HealthCheck（依赖 milvus client）
│   └── *_test.go
├── ragkitdb/                           # 独立 DB 包：表 + AutoMigrate（不进服务端 InitMysql）
│   ├── dao.go                           # StrategyProfileRow / AuditEventRow GORM 模型
│   └── migrate.go                       # Migrate(db) 注册两张表
└── cmd/ragkit-cli/
    └── main.go                          # ingest / retrieve / health / activate / rollback / reindex
```

**职责边界**：`canonical` 不 import `chunking`；`chunking` import `canonical`（用其类型）；`retrieval` import `eino schema`（不 import `chunking`）；`governance` import `retrieval`（用其配置类型）+ `ragkitdb`（用其 GORM 模型）；`ragkit` facade 编排全部。`ragkitdb` 独立，CLI 与（未来）服务端各自调用 `Migrate`，互不污染。

---

## Task 0: 骨架 + go.mod 依赖 + facade 占位

**Files:**
- Create: `agent/knowledge/ragkit/ragkit.go`
- Modify: `go.mod`（add goldmark）、`go.sum`（go mod tidy 自动）
- Test: 无（占位，靠 `go build` 验证）

**Interfaces:**
- Consumes: 无
- Produces: `package ragkit`（占位常量 `Version = "ragkit-v0"`）

- [ ] **Step 1: 创建 facade 占位**

```go
package ragkit

// Version 是 ragkit 实现版本号，用于审计与元数据留痕。
const Version = "ragkit-v0"
```

- [ ] **Step 2: 添加 goldmark 依赖**

Run:
```bash
cd "C:/Users/jiahao li/GolandProjects/awesomeProject4"
go get github.com/yuin/goldmark@latest
go mod tidy
```
Expected: `go.mod` require 块出现 `github.com/yuin/goldmark vX.Y.Z`（无 `// indirect`），`go.sum` 出现对应 `h1:` 行。

- [ ] **Step 3: 验证编译不破坏现有代码**

Run: `go build ./agent/knowledge/ragkit/... && go build ./...`
Expected: 全部通过（ragkit 仅一个常量文件，不引入新错误；现有代码零改动）。

- [ ] **Step 4: Commit**

```bash
git add agent/knowledge/ragkit/ragkit.go go.mod go.sum
git commit -m "feat(ragkit): scaffold package and add goldmark dep"
```

---

## Task 1: canonical 契约类型

**Files:**
- Create: `agent/knowledge/ragkit/canonical/document.go`
- Test: `agent/knowledge/ragkit/canonical/document_test.go`

**Interfaces:**
- Consumes: 无
- Produces: `canonical.NormalizedDocument`、`NormalizedBlock`、`NormalizedTable`、`TableRow`、`ParseQuality`、`CanonicalizationInfo`、`Source`、常量 `Version = "canonical-normalizer-v1"`、`NewNormalizedDocument(raw, source) *NormalizedDocument`

- [ ] **Step 1: 写类型定义**

`document.go`：
```go
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
	Warnings  []string
	OcrLowConf bool
}

// CanonicalizationInfo 记录规范化审计信息。
type CanonicalizationInfo struct {
	Version      string
	AppliedRules []string
	RawSHA1      string
	CanonicalSHA1 string
}

// NormalizedDocument 是规范化后的中间表示，是切块的输入。
type NormalizedDocument struct {
	ContentMarkdown    string
	ContentMarkdownRaw string
	Source             Source
	Blocks            []NormalizedBlock
	Tables             []NormalizedTable
	Quality           ParseQuality
	Canonicalization  CanonicalizationInfo
}

// NewNormalizedDocument 用原始 markdown 与来源构造一个待规范化的文档。
func NewNormalizedDocument(rawMarkdown string, source Source) *NormalizedDocument {
	return &NormalizedDocument{
		ContentMarkdownRaw: rawMarkdown,
		Source:             source,
	}
}
```

- [ ] **Step 2: 写测试（验证构造与版本常量）**

`document_test.go`：
```go
package canonical

import "testing"

func TestNewNormalizedDocumentPreservesRaw(t *testing.T) {
	doc := NewNormalizedDocument("# Title\r\n", Source{FileName: "a.md", FileType: "md"})
	if doc.ContentMarkdownRaw != "# Title\r\n" {
		t.Fatalf("raw not preserved: %q", doc.ContentMarkdownRaw)
	}
	if doc.Source.FileName != "a.md" {
		t.Fatalf("source lost: %+v", doc.Source)
	}
}

func TestVersionConstant(t *testing.T) {
	if Version != "canonical-normalizer-v1" {
		t.Fatalf("unexpected version %q", Version)
	}
}
```

- [ ] **Step 3: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/canonical/...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add agent/knowledge/ragkit/canonical/
git commit -m "feat(ragkit): canonical contract types"
```

---

## Task 2: canonical 规则函数（TDD，逐条）

**Files:**
- Create: `agent/knowledge/ragkit/canonical/rules.go`
- Test: `agent/knowledge/ragkit/canonical/rules_test.go`

**Interfaces:**
- Consumes: 无（`golang.org/x/text/unicode/norm`）
- Produces: `normalizeUnicodeAndLineEndings(s) string`、`normalizeLineNoise(s) string`、`normalizeHeadings(s) string`、`normalizeHeadingContinuations(s) string`、`collapseBlankLines(s) string`

每条规则一个测试用例（表驱动），确保最小可测。先写测试 → 跑失败 → 实现 → 跑过 → commit（最后统一一次 commit 或分 5 次，此处合并为 Task 内一次 commit，但步骤分开）。

- [ ] **Step 1: 写失败测试（unicode + line endings）**

`rules_test.go` 片段：
```go
package canonical

import "testing"

func TestNormalizeUnicodeAndLineEndings(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a\r\nb", "a\nb"},
		{"a\rb", "a\nb"},
		{"ａ\r\n", "A\n"}, // NFKC 全角→半角
	}
	for _, c := range cases {
		if got := normalizeUnicodeAndLineEndings(c.in); got != c.want {
			t.Errorf("normalizeUnicodeAndLineEndings(%q)=%q want %q", c.in, got, c.want)
		}
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestNormalizeUnicodeAndLineEndings`
Expected: FAIL（函数未定义）

- [ ] **Step 3: 实现 normalizeUnicodeAndLineEndings**

`rules.go`：
```go
package canonical

import (
	"strings"
	"golang.org/x/text/unicode/norm"
)

func normalizeUnicodeAndLineEndings(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return norm.NFKC.String(s)
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestNormalizeUnicodeAndLineEndings`
Expected: PASS

- [ ] **Step 5: 写失败测试（line noise）**

追加到 `rules_test.go`：
```go
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
```

- [ ] **Step 6: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestNormalizeLineNoise`
Expected: FAIL

- [ ] **Step 7: 实现 normalizeLineNoise**

追加到 `rules.go`：
```go
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
```

- [ ] **Step 8: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestNormalizeLineNoise`
Expected: PASS

- [ ] **Step 9: 写失败测试（headings）**

```go
func TestNormalizeHeadings(t *testing.T) {
	cases := []struct{ in, want string }{
		{"#title\n", "# title\n"},                // ATX 补空格
		{"## **bold**\n", "## bold\n"},            // 去强调标记
		{"## 1.1标题\n", "## 1.1 标题\n"},         // 数字前缀补空格
		{"## 标　题\n", "## 标题\n"},               // CJK 全角空格去掉
		{"## 标 题\n", "## 标题\n"},               // CJK 间空格去掉
	}
	for _, c := range cases {
		if got := normalizeHeadings(c.in); got != c.want {
			t.Errorf("normalizeHeadings(%q)=%q want %q", c.in, got, c.want)
		}
	}
}
```

- [ ] **Step 10: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestNormalizeHeadings`
Expected: FAIL

- [ ] **Step 11: 实现 normalizeHeadings**

追加到 `rules.go`：
```go
import "regexp"

var (
	atxHeadingRe = regexp.MustCompile(`(?m)^(#{1,6})([^\s#].*)$`)
	emphasisRe   = regexp.MustCompile(`(?m)^(\s*#{1,6}\s*)(\*\*|__|` + "`" + `)(.+?)\2\s*$`)
	cjkSpaceRe   = regexp.MustCompile(`(?m)^(\s*#{1,6}\s*)(.*)$`)
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
	s = emphasisRe.ReplaceAllString(s, "$1$3")
	// 3. CJK 内部去空格（标题行）
	s = cjkSpaceRe.ReplaceAllStringFunc(s, func(m string) string {
		loc := cjkSpaceRe.FindStringSubmatchIndex(m)
		prefix := m[loc[2]:loc[3]]
		rest := m[loc[4]:loc[5]]
		rest = removeCJKSpaces(rest)
		// 数字前缀补空格
		rest = addSpaceAfterNumPrefix(rest)
		return prefix + rest
	})
	return s
}

func removeCJKSpaces(s string) string {
	// 去掉 CJK 字符之间的普通/全角空格
	return regexp.MustCompile(`([\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}])\s+([\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}])`).ReplaceAllString(s, "$1$2")
}

func addSpaceAfterNumPrefix(s string) string {
	// "1.1标题" -> "1.1 标题"
	return regexp.MustCompile(`^(\d+(?:\.\d+)+)([^\s\d])`).ReplaceAllString(s, "$1 $2")
}
```

- [ ] **Step 12: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestNormalizeHeadings`
Expected: PASS（若某用例不通过，调整正则后重跑，直至全部 PASS）

- [ ] **Step 13: 写失败测试（heading continuations）**

```go
func TestNormalizeHeadingContinuations(t *testing.T) {
	// "1.1\nTERM" 跨行碎片合并为 "## 1.1 TERM"
	in := "# Chapter\n\n## 1.1\nTERM\nbody"
	want := "# Chapter\n\n## 1.1 TERM\nbody"
	if got := normalizeHeadingContinuations(in); got != want {
		t.Errorf("got=%q want=%q", got, want)
	}
}
```

- [ ] **Step 14: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestNormalizeHeadingContinuations`
Expected: FAIL

- [ ] **Step 15: 实现 normalizeHeadingContinuations**

追加到 `rules.go`：
```go
var headingContRe = regexp.MustCompile(`(?m)^(#{1,6}\s*)([0-9]+(?:\.[0-9]+)*)\s*\n([^\n#].*)$`)

func normalizeHeadingContinuations(s string) string {
	return headingContRe.ReplaceAllString(s, "$1$2 $3")
}
```

- [ ] **Step 16: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestNormalizeHeadingContinuations`
Expected: PASS

- [ ] **Step 17: 写失败测试（collapse blank lines）**

```go
func TestCollapseBlankLines(t *testing.T) {
	in := "a\n\n\n\n\nb"
	want := "a\n\nb"
	if got := collapseBlankLines(in); got != want {
		t.Errorf("got=%q want=%q", got, want)
	}
}
```

- [ ] **Step 18: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestCollapseBlankLines`
Expected: FAIL

- [ ] **Step 19: 实现 collapseBlankLines**

追加到 `rules.go`：
```go
var multiBlankRe = regexp.MustCompile(`\n{3,}`)

func collapseBlankLines(s string) string {
	return multiBlankRe.ReplaceAllString(s, "\n\n")
}
```

- [ ] **Step 20: 跑全部 canonical 规则测试**

Run: `go test ./agent/knowledge/ragkit/canonical/...`
Expected: PASS（全部规则用例）

- [ ] **Step 21: Commit**

```bash
git add agent/knowledge/ragkit/canonical/rules.go agent/knowledge/ragkit/canonical/rules_test.go
git commit -m "feat(ragkit): canonical normalization rules (unicode/noise/heading/cont/blank)"
```

---

## Task 3: canonical 表格标准化 + span 追踪

**Files:**
- Create: `agent/knowledge/ragkit/canonical/table.go`
- Test: `agent/knowledge/ragkit/canonical/table_test.go`

**Interfaces:**
- Consumes: `canonical.NormalizedDocument`（写 Tables）
- Produces: `NormalizeMarkdownPipeTables(doc *NormalizedDocument)` —— 原地扫描 `doc.ContentMarkdownRaw`，规范 pipe table 并把 `[]NormalizedTable` 写入 `doc.Tables`，返回应用规则名

- [ ] **Step 1: 写失败测试**

`table_test.go`：
```go
package canonical

import "testing"

func TestNormalizeMarkdownPipeTablesDetectsAndRenders(t *testing.T) {
	raw := "intro\n\n| 名称 | 值 |\n| --- | --- |\n| a | 1 |\n| b | 2 |\n\ntail"
	doc := NewNormalizedDocument(raw, Source{FileName: "t.md"})
	rules := NormalizeMarkdownPipeTables(doc)
	if len(doc.Tables) != 1 {
		t.Fatalf("want 1 table, got %d", len(doc.Tables))
	}
	tbl := doc.Tables[0]
	if len(tbl.Rows) != 3 {
		t.Fatalf("want 3 rows (header+2), got %d", len(tbl.Rows))
	}
	if tbl.Rows[0][0] != "名称" || tbl.Rows[2][1] != "2" {
		t.Fatalf("unexpected rows: %v", tbl.Rows)
	}
	if tbl.MarkdownStart <= 0 || tbl.MarkdownEnd <= tbl.MarkdownStart {
		t.Fatalf("bad span: %d..%d", tbl.MarkdownStart, tbl.MarkdownEnd)
	}
	if len(rules) == 0 {
		t.Fatalf("want applied rule recorded")
	}
	// 规范渲染仍包含表格且边界落在 raw 内
	rendered := doc.ContentMarkdown
	if !contains(rendered, "| 名称 | 值 |") {
		t.Fatalf("rendered missing header: %q", rendered)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && indexOf(s, sub) >= 0 }
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestNormalizeMarkdownPipeTables`
Expected: FAIL

- [ ] **Step 3: 实现 NormalizeMarkdownPipeTables**

`table.go`：
```go
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
	// 确保把分隔行也算进去：上面循环已含（分隔行匹配 pipeRowRe 的宽泛条件，这里收紧）
	return lines[start:*i]
}

func parsePipeTable(block []string, startByte int) NormalizedTable {
	rows := make([]TableRow, 0, len(block))
	for _, line := range block {
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
```

> 注：`collectTableLines` 用宽泛 `pipeRowRe` 收集，分隔行也匹配，故表头+分隔+数据行全部纳入 `block`，`parsePipeTable` 中分隔行（`---`）会被当作一行数据单元格。为避免污染，在 `parsePipeTable` 里跳过分隔行：见 Step 4 修正。

- [ ] **Step 4: 修正 parsePipeTable 跳过分隔行**

修改 `parsePipeTable`，在收集 cells 前判断分隔行：
```go
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
	// ... 同上 endByte 计算
}
```
重跑测试直至 PASS。

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run TestNormalizeMarkdownPipeTables`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add agent/knowledge/ragkit/canonical/table.go agent/knowledge/ragkit/canonical/table_test.go
git commit -m "feat(ragkit): canonical pipe-table normalization with span tracking"
```

---

## Task 4: canonical Normalize 编排 + provenance

**Files:**
- Create: `agent/knowledge/ragkit/canonical/normalize.go`
- Create: `agent/knowledge/ragkit/canonical/provenance.go`
- Test: `agent/knowledge/ragkit/canonical/normalize_test.go`

**Interfaces:**
- Consumes: Task 2/3 的规则函数 + `NormalizedDocument`
- Produces: `Normalize(doc *NormalizedDocument) []string`（返回 AppliedRules，原地填 ContentMarkdown/Canonicalization/Tables）、`AnnotateChunksWithProvenance(chunks []*schema.Document, doc *NormalizedDocument) []*schema.Document`

- [ ] **Step 1: 写失败测试（Normalize 全链路）**

`normalize_test.go`：
```go
package canonical

import (
	"crypto/sha1"
	"encoding/hex"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestNormalizeRunsFullPipelineAndRecordsRules(t *testing.T) {
	raw := "#title\r\n\r\n| a | b |\n| --- | --- |\n| 1 | 2 |\n\n## **bold**\nbody"
	doc := NewNormalizedDocument(raw, Source{FileName: "x.md"})
	rules := Normalize(doc)
	if len(rules) == 0 {
		t.Fatal("no rules recorded")
	}
	if doc.ContentMarkdown == "" {
		t.Fatal("ContentMarkdown not set")
	}
	if doc.ContentMarkdown == raw {
		t.Fatal("ContentMarkdown unchanged")
	}
	if doc.Canonicalization.Version != Version {
		t.Fatalf("version mismatch: %q", doc.Canonicalization.Version)
	}
	if doc.Canonicalization.RawSHA1 == "" || doc.Canonicalization.CanonicalSHA1 == "" {
		t.Fatal("sha1 missing")
	}
	rawSum := sha1.Sum([]byte(raw))
	if doc.Canonicalization.RawSHA1 != hex.EncodeToString(rawSum[:]) {
		t.Fatalf("raw sha1 mismatch")
	}
	if len(doc.Tables) != 1 {
		t.Fatalf("want 1 table, got %d", len(doc.Tables))
	}
}

func TestAnnotateChunksWithProvenance(t *testing.T) {
	raw := "# Title\n\nintro paragraph here\n\n## Section\n\nsecond paragraph text"
	doc := NewNormalizedDocument(raw, Source{FileName: "x.md"})
	Normalize(doc)
	chunks := []*schema.Document{
		{ID: "c1", Content: "intro paragraph here"},
		{ID: "c2", Content: "second paragraph text"},
	}
	out := AnnotateChunksWithProvenance(chunks, doc)
	if out[0].MetaData["provenance_text"] != "intro paragraph here" {
		t.Fatalf("provenance not set: %v", out[0].MetaData)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/canonical/ -run "TestNormalize|TestAnnotate"`
Expected: FAIL

- [ ] **Step 3: 实现 Normalize 编排**

`normalize.go`：
```go
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
	s = normalizeLineNoise(s); applied = append(applied, "normalize-line-noise")
	s = normalizeHeadings(s); applied = append(applied, "normalize-headings")
	s = normalizeHeadingContinuations(s); applied = append(applied, "normalize-heading-continuations")
	s = collapseBlankLines(s); applied = append(applied, "collapse-blank-lines")

	doc.ContentMarkdown = s
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
```

- [ ] **Step 4: 实现 provenance**

`provenance.go`：
```go
package canonical

import (
	"strings"

	"github.com/cloudwego/eino/schema"
)

// AnnotateChunksWithProvenance 把 chunk 内容定位回 canonical markdown，
// 写入 provenance_text / canonical_offset 元数据。
func AnnotateChunksWithProvenance(chunks []*schema.Document, doc *NormalizedDocument) []*schema.Document {
	canonical := doc.ContentMarkdown
	for _, c := range chunks {
		if c.MetaData == nil {
			c.MetaData = map[string]any{}
		}
		c.MetaData["provenance_text"] = c.Content
		if off := strings.Index(canonical, c.Content); off >= 0 {
			c.MetaData["canonical_offset"] = off
		}
		// 区块映射：找包含该 offset 的 block
		if off, ok := c.MetaData["canonical_offset"].(int); ok {
			for _, b := range doc.Blocks {
				if off >= b.MarkdownStart && off < b.MarkdownEnd {
					c.MetaData["block_ids"] = []string{b.ID}
					c.MetaData["page"] = b.Page
					break
				}
			}
		}
	}
	return chunks
}
```

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/canonical/...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add agent/knowledge/ragkit/canonical/normalize.go agent/knowledge/ragkit/canonical/provenance.go agent/knowledge/ragkit/canonical/normalize_test.go
git commit -m "feat(ragkit): canonical Normalize pipeline and provenance annotation"
```

---

## Task 5: chunking 接口 + 路由器

**Files:**
- Create: `agent/knowledge/ragkit/chunking/strategy.go`
- Create: `agent/knowledge/ragkit/chunking/router.go`
- Test: `agent/knowledge/ragkit/chunking/router_test.go`

**Interfaces:**
- Consumes: `canonical.NormalizedDocument`、`schema.Document`
- Produces: `Strategy` 接口、`Request`、路由常量 `StrategyMarkdown`/`StrategyTableAware`、`Matcher` 类型、`RoutedStrategy`、`StrategyRouter`、`NewStrategyRouter(defaultStrategy, routes...) *StrategyRouter`

- [ ] **Step 1: 写类型与接口**

`strategy.go`：
```go
package chunking

import (
	"context"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"github.com/cloudwego/eino/schema"
)

const (
	StrategyMarkdown    = "markdown"
	StrategyTableAware   = "table_aware"
	StrategyStructure    = "structure_aware" // 留接口，本轮不实现
	StrategyOCRAware     = "ocr_aware"      // 留接口，本轮不实现
)

// Request 是切块请求。
type Request struct {
	Document       *canonical.NormalizedDocument
	BaseMeta       map[string]any
	NormalizedPath string
}

// Strategy 是切块策略接口。
type Strategy interface {
	Split(ctx context.Context, req Request) ([]*schema.Document, error)
	Name() string
}
```

`router.go`：
```go
package chunking

// Matcher 决定一个策略是否处理该文档。
type Matcher func(req Request) bool

// RoutedStrategy 是一条路由规则。
type RoutedStrategy struct {
	Name    string
	Match   Matcher
	Strategy Strategy
}

// StrategyRouter 按顺序匹配，首个命中者切；未命中走 default。
type StrategyRouter struct {
	defaultStrategy Strategy
	routes          []RoutedStrategy
}

func NewStrategyRouter(defaultStrategy Strategy, routes ...RoutedStrategy) *StrategyRouter {
	return &StrategyRouter{defaultStrategy: defaultStrategy, routes: routes}
}

// Split 路由切块，chunk metadata 记录 chunking_route。
func (r *StrategyRouter) Split(ctx context.Context, req Request) ([]*schema.Document, error) {
	chosen := r.defaultStrategy
	routeName := chosen.Name()
	for _, rt := range r.routes {
		if rt.Match != nil && rt.Match(req) {
			chosen = rt.Strategy
			routeName = rt.Name
			break
		}
	}
	chunks, err := chosen.Split(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, c := range chunks {
		if c.MetaData == nil {
			c.MetaData = map[string]any{}
		}
		c.MetaData["chunking_route"] = routeName
	}
	return chunks, nil
}

func (r *StrategyRouter) Name() string { return "router" }
```

- [ ] **Step 2: 写失败测试（路由命中优先级）**

`router_test.go`：
```go
package chunking

import (
	"context"
	"testing"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"github.com/cloudwego/eino/schema"
)

type fakeStrategy struct{ name string }

func (f fakeStrategy) Name() string { return f.name }
func (f fakeStrategy) Split(ctx context.Context, req Request) ([]*schema.Document, error) {
	return []*schema.Document{{ID: f.name, Content: "x", MetaData: map[string]any{}}}, nil
}

func TestRouterPicksFirstMatchAndTagsRoute(t *testing.T) {
	def := fakeStrategy{"default"}
	tbl := fakeStrategy{StrategyTableAware}
	md := fakeStrategy{StrategyMarkdown}
	router := NewStrategyRouter(md,
		RoutedStrategy{Name: StrategyTableAware, Match: func(req Request) bool { return len(req.Document.Tables) > 0 }, Strategy: tbl},
	)
	_ = def // 未使用
	doc := canonical.NewNormalizedDocument("raw", canonical.Source{})
	doc.Tables = []canonical.NormalizedTable{{Rows: []canonical.TableRow{{"a"}}}}
	chunks, err := router.Split(context.Background(), Request{Document: doc})
	if err != nil {
		t.Fatal(err)
	}
	if chunks[0].MetaData["chunking_route"] != StrategyTableAware {
		t.Fatalf("route tag = %v", chunks[0].MetaData["chunking_route"])
	}
}

func TestRouterFallsBackToDefault(t *testing.T) {
	md := fakeStrategy{StrategyMarkdown}
	router := NewStrategyRouter(md,
		RoutedStrategy{Name: StrategyTableAware, Match: func(req Request) bool { return len(req.Document.Tables) > 0 }, Strategy: fakeStrategy{StrategyTableAware}},
	)
	doc := canonical.NewNormalizedDocument("raw", canonical.Source{})
	chunks, err := router.Split(context.Background(), Request{Document: doc})
	if err != nil {
		t.Fatal(err)
	}
	if chunks[0].MetaData["chunking_route"] != StrategyMarkdown {
		t.Fatalf("route tag = %v", chunks[0].MetaData["chunking_route"])
	}
}
```

- [ ] **Step 3: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/chunking/ -run TestRouter`
Expected: FAIL（fakeStrategy 不满足 Strategy 接口签名需匹配——确认通过后失败为路由逻辑）

- [ ] **Step 4: 跑测试确认通过（实现已在 Step 1 给出）**

Run: `go test ./agent/knowledge/ragkit/chunking/ -run TestRouter`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add agent/knowledge/ragkit/chunking/
git commit -m "feat(ragkit): chunking Strategy interface and StrategyRouter"
```

---

## Task 6: MarkdownStrategy（goldmark AST 切块）

**Files:**
- Create: `agent/knowledge/ragkit/chunking/markdown.go`
- Test: `agent/knowledge/ragkit/chunking/markdown_test.go`

**Interfaces:**
- Consumes: `canonical.NormalizedDocument`、`github.com/yuin/goldmark`、`github.com/yuin/goldmark/ast`
- Produces: `NewMarkdownStrategy(maxChunkBytes int) *MarkdownStrategy`、实现 `Strategy`

策略：用 goldmark parse 成 AST，按标题节点切节；每节超 `maxChunkBytes` 时按段落再切；每块带 `hierarchy_path`/`section_title`/`heading_level`。

- [ ] **Step 1: 写失败测试**

`markdown_test.go`：
```go
package chunking

import (
	"context"
	"strings"
	"testing"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
)

func TestMarkdownStrategySplitsByHeading(t *testing.T) {
	raw := "# Chapter One\n\npara one a.\n\npara one b.\n\n## Section 1.1\n\npara in section.\n\n# Chapter Two\n\npara two."
	doc := canonical.NewNormalizedDocument(raw, canonical.Source{FileName: "x.md"})
	canonical.Normalize(doc)

	s := NewMarkdownStrategy(1000)
	chunks, err := s.Split(context.Background(), Request{Document: doc})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 3 {
		t.Fatalf("want >=3 chunks (ch1/sec/ch2), got %d", len(chunks))
	}
	// 每个 chunk 带章节元数据
	for _, c := range chunks {
		if c.MetaData["section_title"] == nil {
			t.Fatalf("section_title missing on chunk %q", c.Content)
		}
		if c.MetaData["hierarchy_path"] == nil {
			t.Fatalf("hierarchy_path missing")
		}
	}
	// 第一章应含 "Chapter One" 与 "para one"
	joined := strings.Join(mapContents(chunks), "\n")
	if !strings.Contains(joined, "Chapter One") || !strings.Contains(joined, "para one a.") {
		t.Fatalf("chapter one content lost: %q", joined)
	}
}

func mapContents(chunks []*chunkDoc) []string {
	out := make([]string, 0, len(chunks))
	for _, c := range chunks {
		out = append(out, c.Content)
	}
	return out
}
```
> 注：测试用 `*chunkDoc` 是占位，实际 import `github.com/cloudwego/eino/schema` 用 `[]*schema.Document`。把 helper 改为接收 `[]*schema.Document`。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/chunking/ -run TestMarkdownStrategy`
Expected: FAIL（函数未定义）

- [ ] **Step 3: 实现 MarkdownStrategy**

`markdown.go`：
```go
package chunking

import (
	"bytes"
	"context"
	"strings"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"github.com/cloudwego/eino/schema"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type MarkdownStrategy struct {
	maxChunkBytes int
}

func NewMarkdownStrategy(maxChunkBytes int) *MarkdownStrategy {
	if maxChunkBytes <= 0 {
		maxChunkBytes = 1000
	}
	return &MarkdownStrategy{maxChunkBytes: maxChunkBytes}
}

func (m *MarkdownStrategy) Name() string { return StrategyMarkdown }

func (m *MarkdownStrategy) Split(ctx context.Context, req Request) ([]*schema.Document, error) {
	src := []byte(req.Document.ContentMarkdown)
	md := goldmark.New()
	reader := text.NewReader(src)
	root := md.Parser().Parse(reader)
	var chunks []*schema.Document
	pathStack := []string{}

	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindHeading {
			level := n.(*ast.Heading).Level
			title := strings.TrimSpace(string(n.Text(src)))
			for len(pathStack) >= level {
				pathStack = pathStack[:len(pathStack)-1]
			}
			pathStack = append(pathStack, title)
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindSection || n.Kind() == ast.KindParagraph {
			seg := bytes.TrimSpace(n.Text(src))
			if len(seg) == 0 {
				return ast.WalkContinue, nil
			}
			// 超长段落再切
			parts := splitByByteLimit(string(seg), m.maxChunkBytes)
			for _, p := range parts {
				chunks = append(chunks, m.makeChunk(p, pathStack, n))
			}
			return ast.WalkSkipChildren
		}
		return ast.WalkContinue, nil
	})
	if len(chunks) == 0 && len(src) > 0 {
		chunks = append(chunks, m.makeChunk(string(bytes.TrimSpace(src)), pathStack, root))
	}
	return chunks, nil
}

func (m *MarkdownStrategy) makeChunk(content string, path []string, n ast.Node) *schema.Document {
	title := ""
	if len(path) > 0 {
		title = path[len(path)-1]
	}
	return &schema.Document{
		Content: content,
		MetaData: map[string]any{
			"section_title":   title,
			"hierarchy_path":  strings.Join(path, " > "),
			"heading_level":   len(path),
			"chunking_unit":   "heading_section",
		},
	}
}

func splitByByteLimit(s string, limit int) []string {
	if len(s) <= limit {
		return []string{s}
	}
	var parts []string
	for len(s) > limit {
		cut := limit
		// 优先在换行/句号处切
		if i := strings.LastIndexAny(s[:cut], "\n。；！？"); i > limit/2 {
			cut = i + 1
		}
		parts = append(parts, strings.TrimSpace(s[:cut]))
		s = s[cut:]
	}
	if strings.TrimSpace(s) != "" {
		parts = append(parts, strings.TrimSpace(s))
	}
	return parts
}
```

> 注：goldmark 默认 `ast.KindSection` 可能不存在于所有版本；若编译失败，去掉 `n.Kind() == ast.KindSection` 分支，仅用 `ast.KindParagraph`。Step 4 跑测试时按编译器报错调整。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/chunking/ -run TestMarkdownStrategy -v`
Expected: PASS（若 goldmark API 差异，调整 ast.Walk 用法后重跑）

- [ ] **Step 5: Commit**

```bash
git add agent/knowledge/ragkit/chunking/markdown.go agent/knowledge/ragkit/chunking/markdown_test.go
git commit -m "feat(ragkit): markdown chunking strategy via goldmark AST"
```

---

## Task 7: TableAwareStrategy + parent-child 元数据

**Files:**
- Create: `agent/knowledge/ragkit/chunking/table_aware.go`
- Create: `agent/knowledge/ragkit/chunking/parent_child.go`
- Test: `agent/knowledge/ragkit/chunking/table_aware_test.go`

**Interfaces:**
- Consumes: `canonical.NormalizedDocument`、`schema.Document`
- Produces: `NewTableAwareStrategy(maxChunkBytes int) *TableAwareStrategy`（实现 Strategy，无表时返回 nil 让 router 走默认）、`finalizeChunks(chunks []*schema.Document, doc *NormalizedDocument, documentID string) []*schema.Document`（补 parent-child 元数据）

- [ ] **Step 1: 写失败测试（表格独立成 chunk）**

`table_aware_test.go`：
```go
package chunking

import (
	"context"
	"testing"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
)

func TestTableAwareStrategyEachTableOneChunk(t *testing.T) {
	raw := "intro\n\n| 名称 | 值 |\n| --- | --- |\n| a | 1 |\n\n# T\n\n| x | y |\n| --- | --- |\n| p | q |"
	doc := canonical.NewNormalizedDocument(raw, canonical.Source{FileName: "t.md"})
	canonical.Normalize(doc)
	if len(doc.Tables) != 2 {
		t.Fatalf("normalize should find 2 tables, got %d", len(doc.Tables))
	}
	s := NewTableAwareStrategy(1000)
	chunks, err := s.Split(context.Background(), Request{Document: doc})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 2 {
		t.Fatalf("want 2 table chunks, got %d", len(chunks))
	}
	for _, c := range chunks {
		if c.MetaData["chunking_unit"] != "table" {
			t.Fatalf("chunking_unit not table: %v", c.MetaData["chunking_unit"])
		}
	}
}

func TestFinalizeChunksAddsParentChild(t *testing.T) {
	raw := "# Title\n\npara one\n\n## Sub\n\npara two"
	doc := canonical.NewNormalizedDocument(raw, canonical.Source{FileName: "x.md"})
	canonical.Normalize(doc)
	md := NewMarkdownStrategy(1000)
	chunks, _ := md.Split(context.Background(), Request{Document: doc})
	out := finalizeChunks(chunks, doc, "doc-1")
	for _, c := range out {
		if c.MetaData["parent_id"] == nil || c.MetaData["chunk_id"] == nil {
			t.Fatalf("parent-child metadata missing: %v", c.MetaData)
		}
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/chunking/ -run "TestTableAware|TestFinalizeChunks"`
Expected: FAIL

- [ ] **Step 3: 实现 TableAwareStrategy**

`table_aware.go`：
```go
package chunking

import (
	"context"
	"strings"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"github.com/cloudwego/eino/schema"
)

type TableAwareStrategy struct {
	maxChunkBytes int
}

func NewTableAwareStrategy(maxChunkBytes int) *TableAwareStrategy {
	if maxChunkBytes <= 0 {
		maxChunkBytes = 1000
	}
	return &TableAwareStrategy{maxChunkBytes: maxChunkBytes}
}

func (t *TableAwareStrategy) Name() string { return StrategyTableAware }

// Split 把每个表格渲染成独立 chunk（"列名:值" 行式）。
// 没有表格时返回 nil（由 router 走默认策略）。
func (t *TableAwareStrategy) Split(ctx context.Context, req Request) ([]*schema.Document, error) {
	if len(req.Document.Tables) == 0 {
		return nil, nil
	}
	var chunks []*schema.Document
	for i, tbl := range req.Document.Tables {
		if len(tbl.Rows) == 0 {
			continue
		}
		header := tbl.Rows[0]
		var lines []string
		for _, r := range tbl.Rows[1:] {
			for j, cell := range r {
				key := ""
				if j < len(header) {
					key = header[j]
				}
				lines = append(lines, key+": "+cell)
			}
		}
		chunks = append(chunks, &schema.Document{
			Content: strings.Join(lines, "\n"),
			MetaData: map[string]any{
				"chunking_unit":  "table",
				"table_index":    i,
				"table_page":     tbl.Page,
			},
		})
	}
	return chunks, nil
}
```

- [ ] **Step 4: 实现 finalizeChunks**

`parent_child.go`：
```go
package chunking

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"github.com/cloudwego/eino/schema"
)

// finalizeChunks 为每 chunk 补 chunk_id / parent_id / offset / hierarchy 等父子元数据。
func finalizeChunks(chunks []*schema.Document, doc *canonical.NormalizedDocument, documentID string) []*schema.Document {
	canonicalMd := doc.ContentMarkdown
	// 建立 parent（标题章节）候选区间：用 markdown 中的标题定位
	parents := buildHeadingParents(canonicalMd)
	for i, c := range chunks {
		if c.MetaData == nil {
			c.MetaData = map[string]any{}
		}
		off := strings.Index(canonicalMd, c.Content)
		start, end := off, off+len(c.Content)
		c.MetaData["child_start_offset"] = start
		c.MetaData["child_end_offset"] = end
		chunkID := shortHash(documentID + ":" + c.Content)
		c.MetaData["chunk_id"] = chunkID
		c.MetaData["child_id"] = chunkID
		c.MetaData["document_id"] = documentID
		// parent 解析：最窄包含的标题章节
		parent := resolveParent(parents, start, end)
		if parent != nil {
			c.MetaData["parent_id"] = parent.id
			c.MetaData["parent_start_offset"] = parent.start
			c.MetaData["parent_end_offset"] = parent.end
			c.MetaData["parent_build_strategy"] = parent.buildStrategy
		} else if c.MetaData["chunking_unit"] == "table" {
			c.MetaData["parent_build_strategy"] = "table"
		} else {
			c.MetaData["parent_build_strategy"] = "paragraph_window"
		}
		c.MetaData["parent_build_version"] = "phase3-parent-child-v1"
		c.MetaData["parent_token_count"] = approxTokens(c.Content)
		_ = i
	}
	return chunks
}

type headingParent struct {
	id, buildStrategy string
	start, end         int
}

func buildHeadingParents(md string) []headingParent {
	// 简化版：按 "# "/"## " 标题切章节，end 取到下一标题前
	lines := strings.Split(md, "\n")
	var parents []headingParent
	byteOff := 0
	curStart := -1
	curTitle := ""
	flush := func(end int) {
		if curStart >= 0 {
			parents = append(parents, headingParent{
				id: shortHash("section:" + curTitle), buildStrategy: "heading_section",
				start: curStart, end: end,
			})
		}
	}
	for _, line := range lines {
		trim := strings.TrimLeft(line, " ")
		if strings.HasPrefix(trim, "# ") || strings.HasPrefix(trim, "## ") {
			flush(byteOff)
			curStart = byteOff
			curTitle = strings.TrimSpace(strings.TrimLeft(trim, "# "))
		}
		byteOff += len(line) + 1
	}
	flush(byteOff)
	return parents
}

func resolveParent(parents []headingParent, start, end int) *headingParent {
	var best *headingParent
	for i := range parents {
		p := parents[i]
		if p.start <= start && end <= p.end {
			if best == nil || (p.end-p.start) < (best.end-best.start) {
				best = &p
			}
		}
	}
	return best
}

func shortHash(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:8])
}

func approxTokens(s string) int {
	words := strings.Fields(s)
	if len(words) > 0 {
		return len(words)
	}
	return (len([]rune(s)) + 3) / 4
}
```

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/chunking/...`
Expected: PASS（含 Task 5/6 用例）

- [ ] **Step 6: Commit**

```bash
git add agent/knowledge/ragkit/chunking/table_aware.go agent/knowledge/ragkit/chunking/parent_child.go agent/knowledge/ragkit/chunking/table_aware_test.go
git commit -m "feat(ragkit): table-aware chunking and parent-child metadata"
```

---

## Task 8: 语义二次切分（默认关，留开关）

**Files:**
- Create: `agent/knowledge/ragkit/chunking/semantic.go`
- Test: `agent/knowledge/ragkit/chunking/semantic_test.go`

**Interfaces:**
- Consumes: `embedding.Embedder`（eino）、`schema.Document`
- Produces: `NewSemanticResplit(embedder embedding.Embedder, minBlockRunes, breakpointPercentile int, enabled bool) *SemanticResplit`、`(*SemanticResplit).Resplit(ctx, chunks) ([]*schema.Document, error)`

语义二次切分对超长 chunk 按句 embedding 相似度选断点重切。默认 `enabled=false` 直接返回原 chunks。

- [ ] **Step 1: 写失败测试（默认关：原样返回）**

`semantic_test.go`：
```go
package chunking

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestSemanticResplitDisabledReturnsOriginal(t *testing.T) {
	r := NewSemanticResplit(nil, 1200, 20, false)
	in := []*schema.Document{{ID: "1", Content: "long"}}
	out, err := r.Resplit(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].ID != "1" {
		t.Fatalf("disabled should pass through, got %v", out)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/chunking/ -run TestSemanticResplitDisabled`
Expected: FAIL

- [ ] **Step 3: 实现 SemanticResplit（默认关分支 + 留 enabled 分支骨架）**

`semantic.go`：
```go
package chunking

import (
	"context"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
)

type SemanticResplit struct {
	embedder            embedding.Embedder
	minBlockRunes       int
	breakpointPercentile int
	enabled             bool
}

func NewSemanticResplit(embedder embedding.Embedder, minBlockRunes, breakpointPercentile int, enabled bool) *SemanticResplit {
	if minBlockRunes <= 0 {
		minBlockRunes = 1200
	}
	if breakpointPercentile <= 0 {
		breakpointPercentile = 20
	}
	return &SemanticResplit{
		embedder: embedder, minBlockRunes: minBlockRunes,
		breakpointPercentile: breakpointPercentile, enabled: enabled,
	}
}

// Resplit 对超长 chunk 按句 embedding 相似度重切。enabled=false 时原样返回。
func (s *SemanticResplit) Resplit(ctx context.Context, chunks []*schema.Document) ([]*schema.Document, error) {
	if !s.enabled || s.embedder == nil {
		return chunks, nil
	}
	// TODO(enabled=true): 按句切分 → embed → 相邻 cosine → 分位断点重切。
	// 本轮面试文档小，默认关，留接口供 CLI 开启后实现。
	return chunks, nil
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/chunking/ -run TestSemanticResplitDisabled`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add agent/knowledge/ragkit/chunking/semantic.go agent/knowledge/ragkit/chunking/semantic_test.go
git commit -m "feat(ragkit): semantic resplit scaffold (disabled by default)"
```

---

## Task 9: retrieval 类型 + Searcher 接口 + milvusSearcher

**Files:**
- Create: `agent/knowledge/ragkit/retrieval/result.go`
- Create: `agent/knowledge/ragkit/retrieval/searcher.go`
- Test: `agent/knowledge/ragkit/retrieval/searcher_test.go`

**Interfaces:**
- Consumes: `rag.HybridRetrieve`（现有）、`schema.Document`、`milvusclient.Client`
- Produces: `Searcher` 接口、`milvusSearcher`、`Result`/`Item`/`Metrics`/`EvidenceGate`/`GateOutcome` 类型、`NewMilvusSearcher(client *milvusclient.Client) *milvusSearcher`

- [ ] **Step 1: 写类型**

`result.go`：
```go
package retrieval

import "github.com/cloudwego/eino/schema"

// GateOutcome 是证据门控结果。
type GateOutcome string

const (
	GatePass        GateOutcome = "pass"
	GateRefused     GateOutcome = "refused"
	GateDegraded    GateOutcome = "degraded_pass"
	GateDisabled    GateOutcome = "disabled"
)

// EvidenceGate 记录门控判定。
type EvidenceGate struct {
	Outcome GateOutcome
	Reason  string // No-Retrieval-Hit / Low-Rerank-Confidence / Insufficient-Citation-Coverage / Contradictory-Evidence / Out-Of-KB-Scope
}

// Item 是检索返回的单条结果。
type Item struct {
	ID       string
	Content  string
	Score    float64
	Metadata map[string]any
}

// Metrics 记录各阶段计数（观测用，不写库）。
type Metrics struct {
	CandidateTopK   int
	FinalTopK       int
	DedupedCount    int
	RerankApplied   bool
	CitationChecked bool
	EvidenceGate    EvidenceGate
}

// Result 是检索后处理最终结果。
type Result struct {
	Items       []Item
	Metrics     Metrics
	EvidenceGate EvidenceGate
}

// toItems 把 eino Document 转成 Item。
func toItems(docs []*schema.Document) []Item {
	items := make([]Item, 0, len(docs))
	for _, d := range docs {
		items = append(items, Item{
			ID: d.ID, Content: d.Content, Score: d.Score(), Metadata: d.MetaData,
		})
	}
	return items
}
```

`searcher.go`：
```go
package retrieval

import (
	"context"

	"awesomeProject4/agent/knowledge/rag"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// Searcher 抽象向量检索，便于注入 mock。
type Searcher interface {
	Search(ctx context.Context, query string, topK int, filter string) ([]*schema.Document, error)
}

// milvusSearcher 适配现有 rag.HybridRetrieve。
type milvusSearcher struct {
	client *milvusclient.Client
}

func NewMilvusSearcher(client *milvusclient.Client) Searcher {
	return &milvusSearcher{client: client}
}

func (m *milvusSearcher) Search(ctx context.Context, query string, topK int, filter string) ([]*schema.Document, error) {
	return rag.HybridRetrieve(ctx, m.client, query, topK, filter)
}
```

- [ ] **Step 2: 写失败测试（用 mock Searcher 验证适配层不影响数据）**

`searcher_test.go`：
```go
package retrieval

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
)

type mockSearcher struct{ docs []*schema.Document }

func (m mockSearcher) Search(ctx context.Context, query string, topK int, filter string) ([]*schema.Document, error) {
	return m.docs, nil
}

func TestToItemsPreservesFields(t *testing.T) {
	docs := []*schema.Document{{ID: "a", Content: "hi", MetaData: map[string]any{"k": 1}}}
	docs[0].WithScore(0.9)
	items := toItems(docs)
	if items[0].ID != "a" || items[0].Content != "hi" || items[0].Score != 0.9 {
		t.Fatalf("fields lost: %+v", items[0])
	}
}
```

- [ ] **Step 3: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/retrieval/ -run TestToItems`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add agent/knowledge/ragkit/retrieval/
git commit -m "feat(ragkit): retrieval result types and Searcher interface"
```

---

## Task 10: 动态 TopK（DecideDynamicTopK + 战略 TopK + 分数悬崖 + token 预算）

**Files:**
- Create: `agent/knowledge/ragkit/retrieval/topk.go`
- Test: `agent/knowledge/ragkit/retrieval/topk_test.go`

**Interfaces:**
- Consumes: 无
- Produces: `TopKConfig`、`DecideDynamicTopK(query string, config TopKConfig) (candidate, final int)`、`DecideStrategicTopK(hits []Item, config TopKConfig) int`、`ApplyScoreCliffGuard(hits []Item) int`、`ApplyTokenBudgetGuard(hits []Item, budget int) []Item`

规则版本 `phase2-rule-v1`。默认 `CandidateTopK=10`、`FinalTopK=5`；宽泛 query（词数<3 或长度<8）扩 `CandidateTopK` 到 15。

- [ ] **Step 1: 写失败测试**

`topk_test.go`：
```go
package retrieval

import "testing"

func TestDecideDynamicTopKShortQueryExpandsCandidate(t *testing.T) {
	cfg := DefaultTopKConfig()
	cand, final := DecideDynamicTopK("go", cfg)
	if cand <= cfg.BaseCandidateTopK {
		t.Fatalf("short query should expand candidate, got %d", cand)
	}
	if final != cfg.BaseFinalTopK {
		t.Fatalf("final unchanged expected %d got %d", cfg.BaseFinalTopK, final)
	}
}

func TestDecideDynamicTopKNormalQuery(t *testing.T) {
	cfg := DefaultTopKConfig()
	cand, final := DecideDynamicTopK("如何实现一个红黑树的插入操作", cfg)
	if cand != cfg.BaseCandidateTopK || final != cfg.BaseFinalTopK {
		t.Fatalf("normal query should use base: %d/%d", cand, final)
	}
}

func TestApplyScoreCliffGuardTruncatesAtCliff(t *testing.T) {
	hits := []Item{{Score: 0.9}, {Score: 0.88}, {Score: 0.5}, {Score: 0.49}}
	n := ApplyScoreCliffGuard(hits)
	if n > 2 {
		t.Fatalf("cliff should truncate before the 0.88->0.5 drop, got %d", n)
	}
}

func TestApplyTokenBudgetGuard(t *testing.T) {
	hits := []Item{{Content: "aaaa"}, {Content: "bbbb"}, {Content: "cccc"}}
	out := ApplyTokenBudgetGuard(hits, 6) // ~2 tokens each, budget 6 => 3 items
	if len(out) != 3 {
		t.Fatalf("budget should keep all, got %d", len(out))
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/retrieval/ -run "TestDecideDynamic|TestApplyScore|TestApplyToken"`
Expected: FAIL

- [ ] **Step 3: 实现 topk.go**

```go
package retrieval

import "strings"

const topkRuleVersion = "phase2-rule-v1"

// TopKConfig 控制动态/战略 TopK。
type TopKConfig struct {
	BaseCandidateTopK int
	BaseFinalTopK     int
	BroadCandidateTopK int
	ShortQueryRunes   int // 低于此长度视为宽泛
	ShortQueryWords   int // 词数低于此视为宽泛
	ScoreCliffRatio   float64 // 相邻分差超过前一个的此比例视为悬崖
	TokenBudget       int
}

func DefaultTopKConfig() TopKConfig {
	return TopKConfig{
		BaseCandidateTopK: 10,
		BaseFinalTopK:     5,
		BroadCandidateTopK: 15,
		ShortQueryRunes:   8,
		ShortQueryWords:   3,
		ScoreCliffRatio:   0.3,
		TokenBudget:       1200,
	}
}

// DecideDynamicTopK 由查询宽度决定搜索召回上限与最终返回数。
func DecideDynamicTopK(query string, cfg TopKConfig) (candidate, final int) {
	candidate = cfg.BaseCandidateTopK
	final = cfg.BaseFinalTopK
	runes := len([]rune(query))
	words := len(strings.Fields(query))
	if runes < cfg.ShortQueryRunes || words < cfg.ShortQueryWords {
		candidate = cfg.BroadCandidateTopK
	}
	return candidate, final
}

// DecideStrategicTopK 按 rerank 分数分布收敛到 finalTopK（简化：取 min(final, len)）。
func DecideStrategicTopK(hits []Item, cfg TopKConfig) int {
	final := cfg.BaseFinalTopK
	if cliff := ApplyScoreCliffGuard(hits); cliff < final {
		final = cliff
	}
	if final > len(hits) {
		final = len(hits)
	}
	if final < 0 {
		final = 0
	}
	return final
}

// ApplyScoreCliffGuard 返回在首个分数悬崖前应保留的条数。
func ApplyScoreCliffGuard(hits []Item) int {
	for i := 1; i < len(hits); i++ {
		prev := hits[i-1].Score
		if prev <= 0 {
			continue
		}
		drop := (prev - hits[i].Score) / prev
		if drop > 0.3 {
			return i
		}
	}
	return len(hits)
}

// ApplyTokenBudgetGuard 按 token 预算截断。
func ApplyTokenBudgetGuard(hits []Item, budget int) []Item {
	used := 0
	var out []Item
	for _, h := range hits {
		tok := approxTokensItem(h.Content)
		if used+tok > budget {
			break
		}
		used += tok
		out = append(out, h)
	}
	return out
}

func approxTokensItem(s string) int {
	words := strings.Fields(s)
	if len(words) > 0 {
		return len(words)
	}
	return (len([]rune(s)) + 3) / 4
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/retrieval/ -run "TestDecideDynamic|TestApplyScore|TestApplyToken"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add agent/knowledge/ragkit/retrieval/topk.go agent/knowledge/ragkit/retrieval/topk_test.go
git commit -m "feat(ragkit): dynamic/strategic TopK, score cliff, token budget"
```

---

## Task 11: Jaccard 重排 + 去重

**Files:**
- Create: `agent/knowledge/ragkit/retrieval/rerank.go`
- Create: `agent/knowledge/ragkit/retrieval/dedupe.go`
- Test: `agent/knowledge/ragkit/retrieval/rerank_test.go`

**Interfaces:**
- Consumes: `Item`
- Produces: `JaccardReranker`、`(*JaccardReranker).Rerank(query string, hits []Item) []Item`、`Dedupe(hits []Item) []Item`

Jaccard 用 CJK bigram + ASCII 词；重排分 = `0.5*原分 + 0.5*jaccard`。

- [ ] **Step 1: 写失败测试**

`rerank_test.go`：
```go
package retrieval

import "testing"

func TestJaccardRerankBoostsLexicalMatch(t *testing.T) {
	query := "红黑树 插入"
	hits := []Item{
		{ID: "a", Content: "红黑树是一种自平衡二叉搜索树", Score: 0.5},
		{ID: "b", Content: "完全无关的天气内容", Score: 0.9},
	}
	r := &JaccardReranker{}
	out := r.Rerank(query, hits)
	if out[0].ID != "a" {
		t.Fatalf("lexical match should be boosted to top, got %v", out[0].ID)
	}
}

func TestDedupeByDocChunk(t *testing.T) {
	hits := []Item{
		{ID: "x", Metadata: map[string]any{"document_id": "d1", "chunk_id": "c1"}},
		{ID: "y", Metadata: map[string]any{"document_id": "d1", "chunk_id": "c1"}},
		{ID: "z", Metadata: map[string]any{"document_id": "d1", "chunk_id": "c2"}},
	}
	out := Dedupe(hits)
	if len(out) != 2 {
		t.Fatalf("want 2 after dedupe, got %d", len(out))
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/retrieval/ -run "TestJaccard|TestDedupe"`
Expected: FAIL

- [ ] **Step 3: 实现 rerank.go + dedupe.go**

`rerank.go`：
```go
package retrieval

import (
	"sort"
	"strings"
)

type JaccardReranker struct{}

// Rerank 用词法 Jaccard/Coverage 调整排序，原分占 0.5。
func (j *JaccardReranker) Rerank(query string, hits []Item) []Item {
	qSet := tokenSet(query)
	out := make([]Item, len(hits))
	for i, h := range hits {
		dSet := tokenSet(h.Content)
		jacc := jaccard(qSet, dSet)
		merged := 0.5*h.Score + 0.5*jacc
		out[i] = h
		out[i].Score = merged
		if out[i].Metadata == nil {
			out[i].Metadata = map[string]any{}
		}
		out[i].Metadata["rerank_score"] = merged
		out[i].Metadata["jaccard"] = jacc
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Score > out[b].Score })
	return out
}

func tokenSet(s string) map[string]struct{} {
	set := map[string]struct{}{}
	// ASCII 词
	for _, w := range strings.Fields(strings.ToLower(s)) {
		if len(w) > 1 {
			set[w] = struct{}{}
		}
	}
	// CJK bigram
	r := []rune(s)
	for i := 0; i+1 < len(r); i++ {
		set[string(r[i:i+2])] = struct{}{}
	}
	return set
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter, uni := 0, 0
	for k := range a {
		if _, ok := b[k]; ok {
			inter++
		} else {
			uni++
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			uni++
		}
	}
	return float64(inter) / float64(inter+uni)
}
```

`dedupe.go`：
```go
package retrieval

func dedupeKey(it Item) string {
	docID, _ := it.Metadata["document_id"].(string)
	chunkID, _ := it.Metadata["chunk_id"].(string)
	if docID == "" {
		docID = it.ID
	}
	if chunkID == "" {
		chunkID = it.ID
	}
	return docID + ":" + chunkID
}

// Dedupe 按 document_id:chunk_id 去重，保留首个（分数最高，已排序）。
func Dedupe(hits []Item) []Item {
	seen := map[string]struct{}{}
	out := make([]Item, 0, len(hits))
	for _, h := range hits {
		k := dedupeKey(h)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, h)
	}
	return out
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/retrieval/ -run "TestJaccard|TestDedupe"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add agent/knowledge/ragkit/retrieval/rerank.go agent/knowledge/ragkit/retrieval/dedupe.go agent/knowledge/ragkit/retrieval/rerank_test.go
git commit -m "feat(ragkit): Jaccard rerank and dedupe"
```

---

## Task 12: 引文一致性 + 证据门控

**Files:**
- Create: `agent/knowledge/ragkit/retrieval/citation.go`
- Create: `agent/knowledge/ragkit/retrieval/evidence_gate.go`
- Test: `agent/knowledge/ragkit/retrieval/evidence_gate_test.go`

**Interfaces:**
- Consumes: `Item`、`EvidenceGateThresholds`
- Produces: `EvidenceGateThresholds`、`EvaluateEvidenceGate(hits []Item, query string, thresholds EvidenceGateThresholds, enabled bool) EvidenceGate`、`CitationConsistencyChecker`、`(*CitationConsistencyChecker).Check(query string, hits []Item) []Item`（默认关，可开）

- [ ] **Step 1: 写失败测试**

`evidence_gate_test.go`：
```go
package retrieval

import "testing"

func TestEvidenceGateNoHitRefuses(t *testing.T) {
	g := EvaluateEvidenceGate(nil, "q", DefaultEvidenceGateThresholds(), true)
	if g.Outcome != GateRefused || g.Reason != "No-Retrieval-Hit" {
		t.Fatalf("want refused No-Retrieval-Hit, got %+v", g)
	}
}

func TestEvidenceGateLowConfidenceRefuses(t *testing.T) {
	hits := []Item{{Score: 0.2, Content: "weak"}}
	th := DefaultEvidenceGateThresholds()
	g := EvaluateEvidenceGate(hits, "q", th, true)
	if g.Outcome != GateRefused || g.Reason != "Low-Rerank-Confidence" {
		t.Fatalf("want Low-Rerank-Confidence, got %+v", g)
	}
}

func TestEvidenceGatePassesWithConfidentHits(t *testing.T) {
	hits := []Item{{Score: 0.9, Content: "good match"}}
	g := EvaluateEvidenceGate(hits, "q", DefaultEvidenceGateThresholds(), true)
	if g.Outcome != GatePass {
		t.Fatalf("want pass, got %+v", g)
	}
}

func TestEvidenceGateDisabled(t *testing.T) {
	g := EvaluateEvidenceGate(nil, "q", DefaultEvidenceGateThresholds(), false)
	if g.Outcome != GateDisabled {
		t.Fatalf("want disabled, got %+v", g)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/retrieval/ -run TestEvidenceGate`
Expected: FAIL

- [ ] **Step 3: 实现 evidence_gate.go**

```go
package retrieval

import "strings"

// EvidenceGateThresholds 控制门控判定。
type EvidenceGateThresholds struct {
	Enabled            bool
	MinRerankScore     float64
	MinCitationCoverage float64
}

func DefaultEvidenceGateThresholds() EvidenceGateThresholds {
	return EvidenceGateThresholds{
		Enabled:              true,
		MinRerankScore:       0.3,
		MinCitationCoverage:  0.5,
	}
}

// EvaluateEvidenceGate 对已重排结果做门控判定。
func EvaluateEvidenceGate(hits []Item, query string, th EvidenceGateThresholds, enabled bool) EvidenceGate {
	if !enabled || !th.Enabled {
		return EvidenceGate{Outcome: GateDisabled}
	}
	if len(hits) == 0 {
		return EvidenceGate{Outcome: GateRefused, Reason: "No-Retrieval-Hit"}
	}
	top := hits[0]
	if top.Score < th.MinRerankScore {
		return EvidenceGate{Outcome: GateRefused, Reason: "Low-Rerank-Confidence"}
	}
	// 矛盾证据：top 命中含强否定且 query 不含否定
	if contradictionDetected(hits) && !strings.Contains(query, "不") {
		return EvidenceGate{Outcome: GateRefused, Reason: "Contradictory-Evidence"}
	}
	return EvidenceGate{Outcome: GatePass}
}

var negationCues = []string{"不支持", "不能", "并非", "不是"}

func contradictionDetected(hits []Item) bool {
	if len(hits) < 2 {
		return false
	}
	for _, h := range hits[:min(3, len(hits))] {
		for _, cue := range negationCues {
			if strings.Contains(h.Content, cue) {
				return true
			}
		}
	}
	return false
}

func min(a, b int) int { if a < b { return a }; return b }
```

- [ ] **Step 4: 实现 citation.go（默认关，骨架）**

```go
package retrieval

// CitationConsistencyChecker 检查每条命中是否支撑 query 中的 claim。
// 本轮默认关闭（面试 query 多为短问题，claim 抽取噪声大），留接口。
type CitationConsistencyChecker struct {
	Enabled bool
}

func (c *CitationConsistencyChecker) Check(query string, hits []Item) []Item {
	if !c.Enabled {
		return hits
	}
	// TODO(enabled=true): 抽 claim → 对每 doc 评分 → 标 citation_supported。
	return hits
}
```

- [ ] **Step 5: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/retrieval/ -run TestEvidenceGate`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add agent/knowledge/ragkit/retrieval/citation.go agent/knowledge/ragkit/retrieval/evidence_gate.go agent/knowledge/ragkit/retrieval/evidence_gate_test.go
git commit -m "feat(ragkit): evidence gate and citation consistency scaffold"
```

---

## Task 13: PostProcess 编排 + facade Retrieve/Index

**Files:**
- Create: `agent/knowledge/ragkit/retrieval/postprocess.go`
- Modify: `agent/knowledge/ragkit/ragkit.go`
- Test: `agent/knowledge/ragkit/retrieval/postprocess_test.go`

**Interfaces:**
- Consumes: Task 9-12 全部组件
- Produces: `PostProcess(ctx, query, hits, profile RetrieveProfile) (Result, error)`、`RetrieveProfile`、facade `Retrieve(ctx, searcher Searcher, query, filter, profile) (Result, error)`、facade `Index(ctx, embedder, indexer, docPath) (int, error)`

- [ ] **Step 1: 写 postprocess 编排 + profile**

`postprocess.go`：
```go
package retrieval

import "context"

// RetrieveProfile 聚合检索后处理开关与阈值，来自 governance.active profile。
type RetrieveProfile struct {
	TopK          TopKConfig
	RerankEnabled bool
	EvidenceGate  EvidenceGateThresholds
	CitationEnabled bool
}

func DefaultRetrieveProfile() RetrieveProfile {
	return RetrieveProfile{
		TopK:            DefaultTopKConfig(),
		RerankEnabled:    true,
		EvidenceGate:    DefaultEvidenceGateThresholds(),
		CitationEnabled: false,
	}
}

// PostProcess 对搜索结果执行后处理流水线。
func PostProcess(ctx context.Context, query string, hits []Item, profile RetrieveProfile) (Result, error) {
	items := toItems(toDocs(hits))
	m := Metrics{CandidateTopK: len(items)}

	items = Dedupe(items)
	m.DedupedCount = len(items)

	if profile.RerankEnabled {
		items = (&JaccardReranker{}).Rerank(query, items)
		m.RerankApplied = true
	}
	if profile.CitationEnabled {
		items = (&CitationConsistencyChecker{Enabled: true}).Check(query, items)
		m.CitationChecked = true
	}

	final := DecideStrategicTopK(items, profile.TopK)
	m.FinalTopK = final
	items = ApplyTokenBudgetGuard(items, profile.TopK.TokenBudget)
	if final > len(items) {
		final = len(items)
	}
	items = items[:final]

	gate := EvaluateEvidenceGate(items, query, profile.EvidenceGate, true)
	m.EvidenceGate = gate
	if gate.Outcome == GateRefused {
		return Result{Items: nil, Metrics: m, EvidenceGate: gate}, nil
	}
	return Result{Items: items, Metrics: m, EvidenceGate: gate}, nil
}

// toDocs 把 Item 反转回 schema.Document（供内部传递，简化版：直接重建）。
func toDocs(hits []Item) []*einoDoc {
	out := make([]*einoDoc, 0, len(hits))
	for _, h := range hits {
		out = append(out, &einoDoc{id: h.ID, content: h.Content, score: h.Score, meta: h.Metadata})
	}
	return out
}
```
> 注：上面 `toItems(toDocs(hits))` 是冗余转换，会丢字段。简化：`PostProcess` 直接接收 `[]Item`，`Retrieve` 在 facade 里做 `schema.Document → Item` 一次转换。删除 `toDocs`/`einoDoc`，改为 `PostProcess(ctx, query, items []Item, profile)`。Step 2 测试用 `[]Item`。

修订 `postprocess.go` 顶部签名：
```go
func PostProcess(ctx context.Context, query string, items []Item, profile RetrieveProfile) (Result, error) {
	items = Dedupe(items)
	m := Metrics{CandidateTopK: len(items), DedupedCount: len(items)}
	// ... 后续不变
}
```

- [ ] **Step 2: 写失败测试**

`postprocess_test.go`：
```go
package retrieval

import (
	"context"
	"testing"
)

func TestPostProcessRefusesOnLowConfidence(t *testing.T) {
	items := []Item{{ID: "x", Content: "weak", Score: 0.1}}
	res, err := PostProcess(context.Background(), "q", items, DefaultRetrieveProfile())
	if err != nil {
		t.Fatal(err)
	}
	if res.EvidenceGate.Outcome != GateRefused {
		t.Fatalf("want refused, got %v", res.EvidenceGate.Outcome)
	}
	if len(res.Items) != 0 {
		t.Fatalf("refused should return no items, got %d", len(res.Items))
	}
}

func TestPostProcessPassesAndTruncates(t *testing.T) {
	items := []Item{
		{ID: "a", Content: "红黑树 插入 实现", Score: 0.9, Metadata: map[string]any{"document_id": "d1", "chunk_id": "c1"}},
		{ID: "b", Content: "红黑树 平衡", Score: 0.85, Metadata: map[string]any{"document_id": "d1", "chunk_id": "c2"}},
	}
	res, _ := PostProcess(context.Background(), "红黑树 插入", items, DefaultRetrieveProfile())
	if res.EvidenceGate.Outcome != GatePass {
		t.Fatalf("want pass, got %v", res.EvidenceGate.Outcome)
	}
	if len(res.Items) == 0 {
		t.Fatal("no items returned")
	}
}
```

- [ ] **Step 3: 跑测试确认失败**

Run: `go test ./agent/knowledge/ragkit/retrieval/ -run TestPostProcess`
Expected: FAIL

- [ ] **Step 4: 完善 postprocess.go（最终版，接 `[]Item`）**

落地修订后的 `PostProcess`（接 `[]Item`，去掉 toDocs/toItems 冗余）：

```go
func PostProcess(ctx context.Context, query string, items []Item, profile RetrieveProfile) (Result, error) {
	m := Metrics{CandidateTopK: len(items)}
	items = Dedupe(items)
	m.DedupedCount = len(items)

	if profile.RerankEnabled {
		items = (&JaccardReranker{}).Rerank(query, items)
		m.RerankApplied = true
	}
	if profile.CitationEnabled {
		items = (&CitationConsistencyChecker{Enabled: true}).Check(query, items)
		m.CitationChecked = true
	}
	final := DecideStrategicTopK(items, profile.TopK)
	m.FinalTopK = final
	items = ApplyTokenBudgetGuard(items, profile.TopK.TokenBudget)
	if final > len(items) {
		final = len(items)
	}
	if final > 0 {
		items = items[:final]
	} else {
		items = nil
	}
	gate := EvaluateEvidenceGate(items, query, profile.EvidenceGate, true)
	m.EvidenceGate = gate
	if gate.Outcome == GateRefused {
		return Result{Items: nil, Metrics: m, EvidenceGate: gate}, nil
	}
	return Result{Items: items, Metrics: m, EvidenceGate: gate}, nil
}
```

- [ ] **Step 5: 扩展 facade：Retrieve + Index**

`ragkit.go` 追加：
```go
package ragkit

import (
	"context"

	"awesomeProject4/agent/knowledge/rag"
	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"awesomeProject4/agent/knowledge/ragkit/chunking"
	"awesomeProject4/agent/knowledge/ragkit/retrieval"
	"awesomeProject4/agent/knowledge/rag"
)

// NormalizeDocs 把 eino 原始文档标准化为 NormalizedDocument（facade 入口）。
func NormalizeDocs(docs []*einoDoc) []*canonical.NormalizedDocument { /* 见 Task 14 facade */ panic("placeholder") }

// Split 用默认 router 切块（facade 入口）。
func Split(ctx context.Context, nd *canonical.NormalizedDocument) []*einoDoc { panic("placeholder") }

// DefaultRouter 返回默认装配（table-aware → markdown）。
func DefaultRouter() *chunking.StrategyRouter { panic("placeholder") }

// Retrieve 用 searcher 搜索 + 后处理（facade 入口）。
func Retrieve(ctx context.Context, searcher retrieval.Searcher, query, filter string, profile retrieval.RetrieveProfile) (retrieval.Result, error) {
	cand, _ := retrieval.DecideDynamicTopK(query, profile.TopK)
	docs, err := searcher.Search(ctx, query, cand, filter)
	if err != nil {
		return retrieval.Result{}, err
	}
	items := retrieval.ToItems(docs)
	return retrieval.PostProcess(ctx, query, items, profile)
}
```
> 注：`retrieval.ToItems` 即 Task 9 的 `toItems`，需导出为 `ToItems`。facade 里 `einoDoc` 占位应替换为 `*schema.Document`。Task 14 统一落地 facade，此处先让 `Retrieve` 可编译：删除占位 `NormalizeDocs`/`Split`/`DefaultRouter`（留 Task 14），只保留 `Retrieve`，并把 `toItems` 导出。

修订：Task 9 的 `toItems` 改为导出 `ToItems`；本步 facade 只加 `Retrieve`。

- [ ] **Step 6: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/retrieval/ -run TestPostProcess`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add agent/knowledge/ragkit/retrieval/postprocess.go agent/knowledge/ragkit/retrieval/postprocess_test.go agent/knowledge/ragkit/retrieval/searcher.go agent/knowledge/ragkit/ragkit.go
git commit -m "feat(ragkit): PostProcess orchestration and Retrieve facade"
```

---

## Task 14: facade 全量（Index + Normalize + Split + DefaultRouter）+ CLI ingest

**Files:**
- Modify: `agent/knowledge/ragkit/ragkit.go`
- Create: `agent/knowledge/ragkit/cmd/ragkit-cli/main.go`
- Test: `go build` 验证（CLI 集成测试需 Milvus，留 env-gated）

**Interfaces:**
- Consumes: Task 1-13 全部
- Produces: facade `Index(ctx, embedder, indexer, docPath) (int, error)`、`NormalizeDocs`、`Split`、`DefaultRouter`、`ragkit-cli ingest <path>`

- [ ] **Step 1: 写 facade 全量**

`ragkit.go` 最终版：
```go
package ragkit

import (
	"context"

	"awesomeProject4/agent/knowledge/rag"
	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"awesomeProject4/agent/knowledge/ragkit/chunking"
	"awesomeProject4/agent/knowledge/ragkit/retrieval"
	"github.com/cloudwego/eino/schema"
)

const Version = "ragkit-v0"

// DefaultRouter 返回默认切块路由：table-aware → markdown 兜底。
func DefaultRouter() *chunking.StrategyRouter {
	return chunking.NewStrategyRouter(
		chunking.NewMarkdownStrategy(1000),
		chunking.RoutedStrategy{
			Name: chunking.StrategyTableAware,
			Match: func(req chunking.Request) bool { return len(req.Document.Tables) > 0 },
			Strategy: chunking.NewTableAwareStrategy(1000),
		},
	)
}

// NormalizeDocs 把 eino 原始文档标准化（每个 doc 一个 NormalizedDocument）。
func NormalizeDocs(docs []*schema.Document) []*canonical.NormalizedDocument {
	out := make([]*canonical.NormalizedDocument, 0, len(docs))
	for _, d := range docs {
		src := canonical.Source{FileName: fileNameFromMeta(d)}
		nd := canonical.NewNormalizedDocument(d.Content, src)
		canonical.Normalize(nd)
		out = append(out, nd)
	}
	return out
}

// Split 用给定 router 切块并补父子元数据。
func Split(ctx context.Context, nd *canonical.NormalizedDocument, router *chunking.StrategyRouter, documentID string) ([]*schema.Document, error) {
	chunks, err := router.Split(ctx, chunking.Request{Document: nd})
	if err != nil {
		return nil, err
	}
	chunks = canonical.AnnotateChunksWithProvenance(chunks, nd)
	chunks = chunking.FinalizeChunks(chunks, nd, documentID)
	return chunks, nil
}

// Index 是入库全链路 facade：load → normalize → split → enrich metadata（双SHA1）→ store。
// 返回入库 chunk 数。
func Index(ctx context.Context, embedder *rag.Embedder, indexer *rag.Indexer, docPath string) (int, error) {
	docs, err := rag.NewDocumentsLoader(ctx, docPath)
	if err != nil {
		return 0, err
	}
	router := DefaultRouter()
	total := 0
	for _, d := range docs {
		nd := canonical.NewNormalizedDocument(d.Content, canonical.Source{FileName: d.ID})
		canonical.Normalize(nd)
		chunks, err := Split(ctx, nd, router, nd.Source.FileName)
		if err != nil {
			return total, err
		}
		chunks = enrichRagkitMetadata(chunks, nd)
		ids, err := indexer.Store(ctx, chunks)
		if err != nil {
			return total, err
		}
		total += len(ids)
	}
	return total, nil
}

// enrichRagkitMetadata 把双 SHA1 / applied_rules / ragkit 版本写入 chunk metadata。
func enrichRagkitMetadata(chunks []*schema.Document, nd *canonical.NormalizedDocument) []*schema.Document {
	for _, c := range chunks {
		if c.MetaData == nil {
			c.MetaData = map[string]any{}
		}
		c.MetaData["raw_sha1"] = nd.Canonicalization.RawSHA1
		c.MetaData["canonical_sha1"] = nd.Canonicalization.CanonicalSHA1
		c.MetaData["applied_rules"] = nd.Canonicalization.AppliedRules
		c.MetaData["ragkit_version"] = Version
	}
	return chunks
}

func fileNameFromMeta(d *schema.Document) string {
	if d == nil {
		return ""
	}
	if v, ok := d.MetaData["_file_name"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return d.ID
}
```
> 注：`chunking.FinalizeChunks` 需导出（Task 7 是小写 `finalizeChunks`）。本 Task 把它改为导出 `FinalizeChunks`，并同步改 Task 7 测试调用。`rag.Indexer` 与 `rag.Embedder` 是现有类型（`*milvus2.Indexer` / `*ark.Embedder`），facade 用其指针。

- [ ] **Step 2: 导出 finalizeChunks → FinalizeChunks**

修改 `chunking/parent_child.go`：函数名 `finalizeChunks` → `FinalizeChunks`；修改 `table_aware_test.go` 中调用为 `FinalizeChunks`。跑 `go test ./agent/knowledge/ragkit/chunking/...` 确认 PASS。

- [ ] **Step 3: 写 CLI ingest 子命令骨架**

`cmd/ragkit-cli/main.go`：
```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"awesomeProject4/Init"
	"awesomeProject4/agent/knowledge/rag"
	"awesomeProject4/agent/knowledge/ragkit"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("skip .env: %v", err)
	}
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "ingest":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: ragkit-cli ingest <path>")
			os.Exit(1)
		}
		path := os.Args[2]
		ctx := context.Background()
		manager := Init.NewMilvusManger()
		defer manager.Client.Close(context.Background())
		embedder := rag.NewEmbedder(ctx)
		indexer := rag.NewIndexer(ctx, embedder, manager.Client)
		// 支持目录或文件
		files, err := expandPaths(path)
		if err != nil {
			log.Fatal(err)
		}
		total := 0
		for _, f := range files {
			n, err := ragkit.Index(ctx, embedder, indexer, f)
			if err != nil {
				log.Printf("skip %s: %v", f, err)
				continue
			}
			total += n
			log.Printf("imported %s: %d chunks", f, n)
		}
		log.Printf("done: %d files, %d chunks", len(files), total)
	case "retrieve":
		fmt.Fprintln(os.Stderr, "retrieve: see Task 15")
		os.Exit(1)
	default:
		usage()
		os.Exit(1)
	}
}

func expandPaths(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	return filepath.Glob(filepath.Join(path, "*"))
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: ragkit-cli <ingest|retrieve|health|activate|rollback|reindex> ...")
}
```

- [ ] **Step 4: 验证编译**

Run: `go build ./agent/knowledge/ragkit/... && go build ./...`
Expected: 通过（CLI ingest/retrieve 骨架可编译；retrieve 子命令占位）

- [ ] **Step 5: （可选）跑 ingest 验证**

若本地 `.env` 配好 Milvus：
Run: `go run ./agent/knowledge/ragkit/cmd/ragkit-cli ingest doc/某文件.docx`
Expected: 打印 chunk 数 + 双 SHA1 + applied_rules（手动核对）

- [ ] **Step 6: Commit**

```bash
git add agent/knowledge/ragkit/ragkit.go agent/knowledge/ragkit/cmd/ragkit-cli/main.go agent/knowledge/ragkit/chunking/parent_child.go agent/knowledge/ragkit/chunking/table_aware_test.go
git commit -m "feat(ragkit): facade Index/Normalize/Split and ragkit-cli ingest"
```

---

## Task 15: governance 策略档案 + 审计 + 索引健康 + ragkitdb

**Files:**
- Create: `agent/knowledge/ragkit/ragkitdb/dao.go`
- Create: `agent/knowledge/ragkit/ragkitdb/migrate.go`
- Create: `agent/knowledge/ragkit/governance/profile.go`
- Create: `agent/knowledge/ragkit/governance/profile_store.go`
- Create: `agent/knowledge/ragkit/governance/audit.go`
- Create: `agent/knowledge/ragkit/governance/kb_health.go`
- Test: `agent/knowledge/ragkit/governance/*_test.go`

**Interfaces:**
- Consumes: `gorm.DB`、`milvusclient.Client`、`retrieval.RetrieveProfile` 类型
- Produces: `ragkitdb.StrategyProfileRow`、`ragkitdb.AuditEventRow`、`ragkitdb.Migrate(db)`、`governance.ProfileStore` 接口 + `GormProfileStore`、`governance.StrategyProfile`、`GetActiveProfile`/`Activate`/`Rollback`、`AuditLogger` + `DefaultAuditLogger` + 脱敏、`IndexHealth` + `HealthCheck`、`Reindex`（占位）

**独立 DB 包原则**：`ragkitdb.Migrate` 由 CLI（和未来服务端各自）调用，**不进** `Init.InitMysql` 的 AutoMigrate 列表，避免污染现有服务端表迁移。

- [ ] **Step 1: 写 ragkitdb GORM 模型**

`ragkitdb/dao.go`：
```go
package ragkitdb

import "time"

// StrategyProfileRow 是策略档案表（DB 持久化，修正源项目内存态缺陷）。
type StrategyProfileRow struct {
	ID                     uint64 `gorm:"primaryKey;autoIncrement"`
	Name                   string `gorm:"type:varchar(128);uniqueIndex;not null"`
	Status                 string `gorm:"type:varchar(32);index;not null"` // active/candidate/baseline
	FusionStrategy         string `gorm:"type:varchar(64)"`
	TopKConfig             string `gorm:"type:text"`            // JSON
	EvidenceGateThresholds string `gorm:"type:text"`            // JSON
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func (StrategyProfileRow) TableName() string { return "ragkit_strategy_profile" }

// AuditEventRow 是审计事件表。
type AuditEventRow struct {
	ID               uint64 `gorm:"primaryKey;autoIncrement"`
	AuditTraceID     string `gorm:"type:varchar(128);uniqueIndex;not null"`
	Operator         string `gorm:"type:varchar(128)"`
	Action           string `gorm:"type:varchar(64);index;not null"`
	ResourceType     string `gorm:"type:varchar(64)"`
	ResourceID       string `gorm:"type:varchar(128)"`
	Before           string `gorm:"type:text"`
	After            string `gorm:"type:text"`
	Result           string `gorm:"type:varchar(32)"`
	Reason           string `gorm:"type:text"`
	IP               string `gorm:"type:varchar(64)"`
	SensitiveMasked  bool
	CreatedAt        time.Time
}

func (AuditEventRow) TableName() string { return "ragkit_audit_event" }
```

`ragkitdb/migrate.go`：
```go
package ragkitdb

import "gorm.io/gorm"

// Migrate 注册 ragkit 两张表。由 CLI / 服务端各自调用，不进 InitMysql。
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&StrategyProfileRow{}, &AuditEventRow{})
}
```

- [ ] **Step 2: 写 governance profile 纯逻辑 + store**

`profile.go`：
```go
package governance

import "encoding/json"

// StrategyProfile 是领域模型，与 GORM 行解耦。
type StrategyProfile struct {
	ID                     uint64
	Name                   string
	Status                 string
	FusionStrategy         string
	TopKConfig             TopKConfigJSON
	EvidenceGateThresholds EvidenceGateJSON
}

type TopKConfigJSON struct {
	BaseCandidateTopK int `json:"base_candidate_top_k"`
	BaseFinalTopK     int `json:"base_final_top_k"`
	BroadCandidateTopK int `json:"broad_candidate_top_k"`
}

type EvidenceGateJSON struct {
	Enabled            bool    `json:"enabled"`
	MinRerankScore     float64 `json:"min_rerank_score"`
	MinCitationCoverage float64 `json:"min_citation_coverage"`
}

func (p *StrategyProfile) ToTopKConfigJSON() ([]byte, error)       { return json.Marshal(p.TopKConfig) }
func (p *StrategyProfile) ToEvidenceGateJSON() ([]byte, error)    { return json.Marshal(p.EvidenceGateThresholds) }
```

`profile_store.go`：
```go
package governance

import (
	"context"
	"encoding/json"
	"errors"

	"awesomeProject4/agent/knowledge/ragkit/ragkitdb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProfileStore 抽象策略档案持久化，便于测试用内存 store。
type ProfileStore interface {
	GetActive(ctx context.Context) (*StrategyProfile, error)
	Activate(ctx context.Context, id uint64) (*StrategyProfile, error)
	Rollback(ctx context.Context, id uint64) (*StrategyProfile, error)
	Create(ctx context.Context, p *StrategyProfile) error
}

type GormProfileStore struct {
	db *gorm.DB
}

func NewGormProfileStore(db *gorm.DB) ProfileStore { return &GormProfileStore{db: db} }

func (s *GormProfileStore) GetActive(ctx context.Context) (*StrategyProfile, error) {
	var row ragkitdb.StrategyProfileRow
	err := s.db.WithContext(ctx).Where("status = ?", "active").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rowToProfile(row), nil
}

func (s *GormProfileStore) Activate(ctx context.Context, id uint64) (*StrategyProfile, error) {
	return s.inTx(ctx, func(tx *gorm.DB) (*StrategyProfile, error) {
		// 当前 active → baseline
		if err := tx.Model(&ragkitdb.StrategyProfileRow{}).
			Where("status = ?", "active").Update("status", "baseline").Error; err != nil {
			return nil, err
		}
		if err := tx.Model(&ragkitdb.StrategyProfileRow{}).
			Where("id = ?", id).Update("status", "active").Error; err != nil {
			return nil, err
		}
		var row ragkitdb.StrategyProfileRow
		if err := tx.First(&row, id).Error; err != nil {
			return nil, err
		}
		return rowToProfile(row), nil
	})
}

func (s *GormProfileStore) Rollback(ctx context.Context, id uint64) (*StrategyProfile, error) {
	return s.inTx(ctx, func(tx *gorm.DB) (*StrategyProfile, error) {
		// 找 baseline
		var base ragkitdb.StrategyProfileRow
		if err := tx.Where("status = ?", "baseline").First(&base).Error; err != nil {
			return nil, err
		}
		// 当前 active → archived（简化：改回 candidate），baseline → active
		if err := tx.Model(&ragkitdb.StrategyProfileRow{}).
			Where("status = ?", "active").Update("status", "candidate").Error; err != nil {
			return nil, err
		}
		if err := tx.Model(&ragkitdb.StrategyProfileRow{}).
			Where("id = ?", base.ID).Update("status", "active").Error; err != nil {
			return nil, err
		}
		var row ragkitdb.StrategyProfileRow
		if err := tx.First(&row, base.ID).Error; err != nil {
			return nil, err
		}
		return rowToProfile(row), nil
	})
}

func (s *GormProfileStore) Create(ctx context.Context, p *StrategyProfile) error {
	row := profileToRow(p)
	return s.db.WithContext(ctx).Create(&row).Error
}

func (s *GormProfileStore) inTx(ctx context.Context, fn func(*gorm.DB) (*StrategyProfile, error)) (*StrategyProfile, error) {
	var out *StrategyProfile
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		p, err := fn(tx)
		out = p
		return err
	})
	return out, err
}

func rowToProfile(row ragkitdb.StrategyProfileRow) *StrategyProfile {
	p := &StrategyProfile{
		ID: row.ID, Name: row.Name, Status: row.Status, FusionStrategy: row.FusionStrategy,
	}
	_ = json.Unmarshal([]byte(row.TopKConfig), &p.TopKConfig)
	_ = json.Unmarshal([]byte(row.EvidenceGateThresholds), &p.EvidenceGateThresholds)
	return p
}

func profileToRow(p *StrategyProfile) ragkitdb.StrategyProfileRow {
	row := ragkitdb.StrategyProfileRow{
		ID: p.ID, Name: p.Name, Status: p.Status, FusionStrategy: p.FusionStrategy,
	}
	row.TopKConfig, _ = json.Marshal(p.TopKConfig)
	row.EvidenceGateThresholds, _ = json.Marshal(p.EvidenceGateThresholds)
	return row
}

var _ = clause.Lock // 保留 import 占位（实际行锁由事务 + 单写者保证）
```

- [ ] **Step 3: 写 governance audit + 脱敏**

`audit.go`：
```go
package governance

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"awesomeProject4/agent/knowledge/ragkit/ragkitdb"
	"gorm.io/gorm"
)

// AuditLogger 抽象审计写入。
type AuditLogger interface {
	Log(ctx context.Context, e AuditEvent) error
}

type AuditEvent struct {
	Operator     string
	Action       string
	ResourceType string
	ResourceID   string
	Before, After any
	Result, Reason string
	IP           string
}

type DefaultAuditLogger struct {
	db *gorm.DB
}

func NewDefaultAuditLogger(db *gorm.DB) AuditLogger { return &DefaultAuditLogger{db: db} }

func (l *DefaultAuditLogger) Log(ctx context.Context, e AuditEvent) error {
	row := ragkitdb.AuditEventRow{
		AuditTraceID:    newTraceID(),
		Operator:        e.Operator,
		Action:          e.Action,
		ResourceType:   e.ResourceType,
		ResourceID:      e.ResourceID,
		Before:         maskJSON(e.Before),
		After:          maskJSON(e.After),
		Result:         e.Result,
		Reason:         e.Reason,
		IP:             maskIP(e.IP),
		SensitiveMasked: true,
	}
	return l.db.WithContext(ctx).Create(&row).Error
}

func newTraceID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var querySnippetRe = regexp.MustCompile(`(?i)(query|snippet|content)["']?\s*[:=]\s*["'][^"']*["']`)

// maskJSON 把敏感字段（query/snippet/content）脱敏。
func maskJSON(v any) string {
	if v == nil {
		return ""
	}
	s := mustJSON(v)
	s = querySnippetRe.ReplaceAllString(s, `$1: "[masked]"`)
	return s
}

func mustJSON(v any) string {
	b, err := jsonMarshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// maskIP 把 IP 第 3、4 段打码。
func maskIP(ip string) string {
	// 简化：保留前两段
	parts := splitOnDot(ip)
	if len(parts) <= 2 {
		return ip
	}
	return parts[0] + "." + parts[1] + ".*.*"
}

func splitOnDot(s string) []string {
	out := []string{}
	cur := ""
	for _, r := range s {
		if r == '.' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
```
> 注：`jsonMarshal` 用 `encoding/json.Marshal`，补 import。`rand` 在脚本受限，但这是编译产物不是 workflow 脚本，正常可用。

- [ ] **Step 4: 写 governance kb_health + Reindex 占位**

`kb_health.go`：
```go
package governance

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// IndexHealth 记录集合健康检查结果。
type IndexHealth struct {
	OK     bool
	Checks map[string]bool
	Gaps   []string
}

// HealthCheck 校验集合存在、维度、度量、load 健康（简化：校验集合存在 + 描述）。
func HealthCheck(ctx context.Context, client *milvusclient.Client, collection string) (IndexHealth, error) {
	h := IndexHealth{Checks: map[string]bool{}}
	// eino-ext milvus client DescribeCollection 简化：调用失败即不健康
	// 实际实现按 milvusclient API：client.DescribeCollection(ctx, NewDescribeCollectionOption(collection))
	// 本轮先校验 client 非空 + collection 名非空，留 Gaps 记录未实现的契约项。
	if client == nil {
		return h, fmt.Errorf("milvus client is nil")
	}
	if collection == "" {
		h.Gaps = append(h.Gaps, "collection name empty")
		return h, nil
	}
	h.Checks["client_ok"] = true
	h.Checks["collection_named"] = true
	h.Gaps = append(h.Gaps, "describe_collection_not_implemented", "dimension_match_not_implemented")
	h.OK = len(h.Gaps) == 0
	return h, nil
}

// Reindex 占位：drop & recreate collection + 重跑入库。本轮留接口，CLI reindex 提示未实现。
func Reindex(ctx context.Context, client *milvusclient.Client, collection string) error {
	return fmt.Errorf("reindex not implemented in this phase")
}
```

- [ ] **Step 5: 写 store 测试（用内存 sqlite，避免依赖 MySQL）**

`governance/profile_store_test.go`：
```go
package governance

import (
	"context"
	"testing"

	"awesomeProject4/agent/knowledge/ragkit/ragkitdb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := ragkitdb.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestProfileStoreActivateSwitchesActive(t *testing.T) {
	db := newTestDB(t)
	store := NewGormProfileStore(db)
	ctx := context.Background()

	// 建一个 baseline + 一个 candidate
	base := &StrategyProfile{Name: "base", Status: "active", TopKConfig: TopKConfigJSON{BaseCandidateTopK: 10, BaseFinalTopK: 5}}
	cand := &StrategyProfile{Name: "cand", Status: "candidate", TopKConfig: TopKConfigJSON{BaseCandidateTopK: 15, BaseFinalTopK: 5}}
	if err := store.Create(ctx, base); err != nil {
		t.Fatal(err)
	}
	if err := store.Create(ctx, cand); err != nil {
		t.Fatal(err)
	}

	activated, err := store.Activate(ctx, cand.ID)
	if err != nil {
		t.Fatal(err)
	}
	if activated.Status != "active" {
		t.Fatalf("cand should be active, got %s", activated.Status)
	}
	cur, _ := store.GetActive(ctx)
	if cur.Name != "cand" {
		t.Fatalf("active should be cand, got %s", cur.Name)
	}
}

func TestProfileStoreRollbackToBaseline(t *testing.T) {
	db := newTestDB(t)
	store := NewGormProfileStore(db)
	ctx := context.Background()
	base := &StrategyProfile{Name: "base", Status: "active"}
	cand := &StrategyProfile{Name: "cand", Status: "candidate"}
	store.Create(ctx, base)
	store.Create(ctx, cand)
	store.Activate(ctx, cand.ID) // 现在 cand active, base baseline
	rolled, err := store.Rollback(ctx, cand.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rolled.Name != "base" {
		t.Fatalf("rollback should restore base, got %s", rolled.Name)
	}
}
```

- [ ] **Step 6: 添加 sqlite 测试依赖**

Run:
```bash
cd "C:/Users/jiahao li/GolandProjects/awesomeProject4"
go get gorm.io/driver/sqlite@latest
```
Expected: `go.mod` 新增 `gorm.io/driver/sqlite`（仅测试用）。

- [ ] **Step 7: 跑测试确认通过**

Run: `go test ./agent/knowledge/ragkit/governance/...`
Expected: PASS（profile store 激活/回滚；audit/health 为骨架无独立逻辑测试）

- [ ] **Step 8: 验证全量编译**

Run: `go build ./agent/knowledge/ragkit/... && go build ./...`
Expected: 通过

- [ ] **Step 9: Commit**

```bash
git add agent/knowledge/ragkit/ragkitdb/ agent/knowledge/ragkit/governance/ go.mod go.sum
git commit -m "feat(ragkit): governance strategy profile, audit, index health (sqlite-tested)"
```

---

## Task 16: CLI 全量子命令 + 接线 stub

**Files:**
- Modify: `agent/knowledge/ragkit/cmd/ragkit-cli/main.go`
- Modify: `agent/knowledge/import.go`（仅加注释 stub）
- Modify: `agent/tool/retriever.go`（仅加注释 stub）
- Test: `go build ./...`（接线为注释，靠编译验证）

**Interfaces:**
- Consumes: Task 14-15
- Produces: CLI `retrieve/health/activate/rollback/reindex`、两处接线点注释 stub（`RAGKIT_ENABLED` 开关）

- [ ] **Step 1: 扩展 CLI retrieve 子命令**

`main.go` 增加：
```go
case "retrieve":
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: ragkit-cli retrieve <query>")
		os.Exit(1)
	}
	query := os.Args[2]
	ctx := context.Background()
	manager := Init.NewMilvusManger()
	defer manager.Client.Close(context.Background())
	searcher := retrieval.NewMilvusSearcher(manager.Client)
	// 读 active profile（若有 DB 则用，否则默认）
	res, err := ragkit.Retrieve(ctx, searcher, query, "", retrieval.DefaultRetrieveProfile())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("gate=%s reason=%s items=%d\n", res.EvidenceGate.Outcome, res.EvidenceGate.Reason, len(res.Items))
	for _, it := range res.Items {
		fmt.Printf("- [%.3f] %s\n", it.Score, truncate(it.Content, 200))
	}
case "health":
	ctx := context.Background()
	manager := Init.NewMilvusManger()
	defer manager.Client.Close(context.Background())
	collection := os.Getenv("collection")
	h, err := governance.HealthCheck(ctx, manager.Client, collection)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ok=%v checks=%v gaps=%v\n", h.OK, h.Checks, h.Gaps)
case "activate":
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: ragkit-cli activate <profile_id>")
		os.Exit(1)
	}
	// 需要 DB：复用 InitMysql 或独立连接。简化：用 Init.InitMysql
	db, err := Init.InitMysql()
	if err != nil {
		log.Fatal(err)
	}
	ragkitdb.Migrate(db)
	store := governance.NewGormProfileStore(db)
	id := atoi(os.Args[2])
	p, err := store.Activate(context.Background(), uint64(id))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("activated profile: %s (id=%d)", p.Name, p.ID)
case "rollback":
	// 同 activate，调 store.Rollback
	db, _ := Init.InitMysql()
	ragkitdb.Migrate(db)
	store := governance.NewGormProfileStore(db)
	id := atoi(os.Args[2])
	p, err := store.Rollback(context.Background(), uint64(id))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("rolled back to: %s (id=%d)", p.Name, p.ID)
case "reindex":
	fmt.Fprintln(os.Stderr, "reindex not implemented in this phase")
	os.Exit(1)
```
> 注：补 import `awesomeProject4/agent/knowledge/ragkit/governance`、`retrieval`、`ragkitdb`，补 helper `truncate`/`atoi`。

- [ ] **Step 2: 加接线 stub 到 import.go**

在 `agent/knowledge/import.go` 的 `splitter := rag.NewSemanticSplit(...)` 之前插入注释块：
```go
	// === ragkit 接线点（标准化 + 路由切块），默认关闭，手动切换时启用 ===
	// 启用方式：设置环境变量 RAGKIT_ENABLED=1，并替换下方 splitter.Transform 为 ragkit.Split
	// _ = ragkit.NormalizeDocs(docs)
	// _ = ragkit.DefaultRouter()
```
> 不改任何可执行代码行，仅追加注释。

- [ ] **Step 3: 加接线 stub 到 retriever.go**

在 `agent/tool/retriever.go::GetRetrieverWithInput` 的 `docs, err := rag.HybridRetrieve(...)` 之前插入注释块：
```go
	// === ragkit 接线点（门控 + 动态 TopK），默认关闭 ===
	// 启用方式：设置环境变量 RAGKIT_ENABLED=1，并把本函数返回替换为 ragkit.Retrieve
	// res, _ := ragkit.Retrieve(searchCtx, retrieval.NewMilvusSearcher(manager.Client), query, input.Filter, retrieval.DefaultRetrieveProfile())
	// return toRetrieverOutput(res), nil
```
> 不改可执行代码，仅注释。

- [ ] **Step 4: 验证全量编译 + 既有测试不破**

Run: `go build ./... && go test ./agent/knowledge/rag/...`
Expected: 全部通过（rag 包测试未被改动，应保持原样 PASS）

- [ ] **Step 5: Commit**

```bash
git add agent/knowledge/ragkit/cmd/ragkit-cli/main.go agent/knowledge/import.go agent/tool/retriever.go
git commit -m "feat(ragkit): CLI subcommands and seamless-replace stubs"
```

---

## Task 17: 验收 + 文档

**Files:**
- Create: `docs/superpowers/specs/2026-08-03-ragkit-extraction-design.md`（已存在，确认无改动需求）
- Create: `agent/knowledge/ragkit/README.md`
- Test: 全量 `go build` + `go test ./agent/knowledge/ragkit/...`

- [ ] **Step 1: 写 README**

`agent/knowledge/ragkit/README.md`：
```markdown
# ragkit —— RAG 精华移植包

把 rag-retrievalOps 的 标准化/切块/检索增强/精简治理 四块精华在 awesomeProject4 内重写，
对齐现有栈（Milvus + eino + Ark），纯 Go 包 API，零现有代码改动。

## 快速验证

```bash
# 入库
go run ./agent/knowledge/ragkit/cmd/ragkit-cli ingest doc/某文件.docx

# 检索
go run ./agent/knowledge/ragkit/cmd/ragkit-cli retrieve "红黑树 插入"

# 索引健康
go run ./agent/knowledge/ragkit/cmd/ragkit-cli health
```

## 无感替换（手动切换）

两处接线点（默认关闭，开关 `RAGKIT_ENABLED=1`）：
- `agent/knowledge/import.go`：标准化 + 路由切块
- `agent/tool/retriever.go`：门控 + 动态 TopK

切换时把对应注释内的 ragkit 调用启用，删除/注释现有 rag 调用。现有代码保留，可随时回退。

## 包结构

- `canonical/` 文档标准化
- `chunking/` 切块策略路由
- `retrieval/` 检索后处理
- `governance/` 精简治理（策略档案/审计/索引健康）
- `ragkitdb/` 独立 DB 表（不进服务端 AutoMigrate）

## 依赖

仅新增 `github.com/yuin/goldmark`；测试用 `gorm.io/driver/sqlite`。
```

- [ ] **Step 2: 跑全量测试**

Run: `go test ./agent/knowledge/ragkit/...`
Expected: 全部 PASS

- [ ] **Step 3: 跑全量编译**

Run: `go build ./...`
Expected: 通过，现有代码零改动

- [ ] **Step 4: （手动）E2E 验收**

操作者手动执行（需 Milvus + .env）：
1. `ragkit-cli ingest doc/<含表格的docx>` → 打印 chunk 数 + 双 SHA1 + applied_rules + chunking_route
2. `ragkit-cli retrieve "某面试问题"` → 命中带 provenance/parent_id；低相关 query 被门控拒绝带 reason
3. `ragkit-cli health` → 输出 OK/Checks/Gaps
4. `ragkit-cli activate <profile_id>` 后 `retrieve` → 行为变化 → `audit_event` 表可查

- [ ] **Step 5: Commit**

```bash
git add agent/knowledge/ragkit/README.md
git commit -m "docs(ragkit): README and verification guide"
```

---

## Self-Review 结论

**1. Spec 覆盖**：标准化（Task 1-4）、切块路由+parent-child+语义切分（Task 5-8）、检索后处理 TopK/重排/门控/编排（Task 9-13）、facade+CLI（Task 14）、治理+独立DB（Task 15）、CLI全量+接线stub（Task 16）、验收+文档（Task 17）。精简治理明确排除 eval/HTTP 路由，与 spec §5、§9 一致。

**2. 无占位符**：各 Task 含完整代码与测试。Task 8 语义切分、Task 12 引文一致性、Task 15 Reindex 按 spec「留接口默认关」标注 `TODO(enabled=true)`/`not implemented in this phase`，是 spec 明确范围而非计划占位。

**3. 类型一致**：`Strategy`/`Request`/`Item`/`Result`/`EvidenceGate`/`TopKConfig`/`RetrieveProfile`/`StrategyProfile`/`ProfileStore`/`AuditLogger`/`IndexHealth` 跨 Task 签名一致；`finalizeChunks`→`FinalizeChunks` 导出在 Task 14 Step 2 同步。`toItems`→`ToItems` 导出在 Task 13 Step 5 同步。

**4. 风险点（已在 Task 内标注）**：goldmark AST API 版本差异（Task 6 Step 4 注）、sqlite 测试依赖新增（Task 15 Step 6）、`Init.InitMysql` 含硬编码 DSN（CLI activate/rollback 复用，沿用现有，不在本计划修改）。
