# RAG 精华移植设计（ragkit）

> 目标：把 rag-retrievalOps 的 **文档标准化 / 切块 / 检索 / 治理** 四块精华，在 `awesomeProject4` 内重写成一套自包含、对齐现有栈（Milvus + eino + Ark）、纯 Go 包 API、可快速验证的实现，做到对现有 RAG 的无感替换。最后切换由用户手动完成。

- 日期：2026-08-03
- 目标仓库：`awesomeProject4`（当前分支 `rag`）
- 来源参考：`rag-retrievalOps`（backend/internal/documentparser、backend/internal/milvus/chunking、backend/internal/milvus/retrieval、backend/internal/milvus/evaluation、backend/api/handler/kb/*）
- 已有同主题计划：`doc/rag-governance-import-plan.md`（本设计与之在范围上有意收窄，见下文「与现有导入计划的关系」）

---

## 0. 决策摘要（来自澄清问答）

| 维度 | 决策 |
|---|---|
| 现状 | awesomeProject4 已有简单 RAG（`agent/knowledge/rag/` + `agent/knowledge/import.go` + `agent/tool/retriever.go`），本次是在其旁写新实现，用户手动切换，需快速验证 |
| 外部依赖 | 对齐 awesomeProject4 现有栈：Milvus（Zilliz）+ eino + Ark `qwen3-embedding-0.6b` 1024 维，不引入新基础设施 |
| 精华范围 | 标准化 / 切块 / 检索 / 治理 四块全要 |
| 治理深度 | 精简版：知识库元数据 + 索引重建 + 索引健康 + 审计事件 + 策略档案（DB 持久化）；**不做** eval 评测、不做 HTTP 治理路由 |
| 暴露形式 | 纯 Go 包 API（`ragkit.Index(...)` / `ragkit.Retrieve(...)`），不新增 HTTP |
| 文档格式 | 首批 Markdown / 纯文本 + PDF；DOCX 走现有 eino loader 不在本包重写解析 |
| 落地方案 | 方案 1：单一 facade 包 `ragkit` + 内部四子包 + 独立验证 CLI |

## 与现有导入计划的关系

`doc/rag-governance-import-plan.md` 已规划过同类工作，但范围更重（含 eval 评测 / HTTP 治理路由 / Kafka eval run / 多表）。本设计按用户「精简治理 + 纯 Go 包 + 快速验证」明确收窄：

- **保留**：标准化、切块路由、检索增强（门控/动态 TopK）、精简治理（审计/策略档案/索引健康）。
- **去掉**：eval 评测 runner、eval 数据集/用例/run 三表、Kafka eval run、`/eval/*` `/strategy/*` `/audit/*` HTTP 路由、minmax/RRF 融合层叠加。
- 融合策略不重做：现有 Milvus 混合检索（dense+sparse+RRF k=60）已足够，本包**不**在 Milvus 之上再叠一层 fusion，只在检索结果上做后处理（去重/TopK/重排/门控）。

---

## 1. 总体架构与包布局

### 1.1 包结构

在 `awesomeProject4` 下新增自包含包 `agent/knowledge/ragkit`，内部四个子包按职责隔离，对外只通过 `ragkit` facade 暴露：

```
agent/knowledge/ragkit/
├── ragkit.go              # facade：Index() / Retrieve() / 等高级 API
├── canonical/             # 文档标准化（NormalizedDocument + 规则链 + 表格 + 双 SHA1）
│   ├── document.go        # NormalizedDocument / Block / Table / Quality / CanonicalizationInfo
│   ├── normalize.go        # Normalize() 入口 + 规则链编排
│   ├── rules.go            # unicode/whitespace / noise / heading / heading-continuation / blank
│   ├── table.go            # NormalizeMarkdownPipeTables + span 追踪
│   └── provenance.go       # AnnotateChunksWithProvenance（chunk 偏移 → block/page）
├── chunking/              # 切块策略路由
│   ├── strategy.go         # Strategy 接口 + Request + 常量
│   ├── router.go           # StrategyRouter（先命中先切 + 路由标签）
│   ├── markdown.go         # MarkdownStrategy（goldmark 标题树 + 字节限长）
│   ├── table_aware.go      # TableAwareStrategy（表格独立成 chunk）
│   ├── parent_child.go     # finalizeChunks 父子元数据补全
│   └── semantic.go         # 语义二次切分（句 embedding 相似度 + 分位断点）
├── retrieval/             # 检索增强（后处理层）
│   ├── searcher.go         # Searcher 接口（适配现有 HybridRetrieve）
│   ├── postprocess.go     # 去重 / 动态 TopK / Jaccard 重排 / 引文一致性 / 证据门控 编排
│   ├── topk.go             # DecideDynamicTopK（查询长度/宽泛度/分数悬崖/证据密度/token 预算）
│   ├── rerank.go           # JaccardReranker（兜底，无外部模型依赖）
│   ├── citation.go         # 引文一致性检查
│   ├── evidence_gate.go    # 证据门控 + 5 类拒绝原因
│   └── dedupe.go           # 去重
├── governance/            # 精简治理
│   ├── knowledge_base.go  # 知识库元数据（命名/集合/嵌入维度）+ 索引重建 + 索引健康
│   ├── strategy_profile.go # 策略档案（DB 持久化：融合/topk/门控阈值）+ active 选择 + 回滚
│   └── audit.go            # 审计事件写入 + 动作常量 + 敏感字段脱敏
└── cmd/
    └── ragkit-cli/         # 独立验证 CLI：ingest / retrieve / reindex / health
        └── main.go
```

### 1.2 设计原则

1. **零现有代码改动**：不修改 `agent/knowledge/rag/`、`agent/knowledge/import.go`、`agent/tool/retriever.go`、`agent/agents/*`、面试 SSE 链路。新包独立存在，用户手动切换。
2. **两处接线点**（仅加 stub，不改逻辑）：
   - `agent/knowledge/import.go`：在 loader 与 splitter 之间可插入 `canonical.Normalize`，splitter 可替换为 `StrategyRouter`。
   - `agent/tool/retriever.go::GetRetrieverWithInput`：可包装 `ragkit.Retrieve`（含门控/动态 TopK）。
   - 接线形式为**可选 stub**（feature flag 或注释占位），默认不启用，切换由用户手动完成。
3. **对齐现有栈**：复用 `Init.NewMilvusManger`、eino `milvus2` indexer/retriever、`ark.NewEmbedder`、`rag.NewBatchedEmbedder`。集合名 `documents`、维度 1024、COSINE+HNSW、BM25 sparse 全部沿用。
4. **依赖最小**：仅新增 `github.com/yuin/goldmark`（Markdown 标题树切块，源项目同款，与 eino 兼容）。
5. **接口隔离**：每个子包单一职责、定义清晰接口，可独立测试。`canonical` 不依赖 chunking；chunking 依赖 canonical 类型；retrieval 依赖 eino schema；governance 依赖 GORM + retrieval 配置类型。

### 1.3 数据流

**入库（Index）**：
```
文件 → eino file loader（pdf/docx/txt/md，复用现有）→ *schema.Document
  → canonical.Normalize → NormalizedDocument（ContentMarkdown + Blocks + Tables + 双SHA1 + AppliedRules）
  → chunking.StrategyRouter.Split → []*schema.Document（带 parent-child 元数据 + 路由标签 + 双SHA1）
  → 现有 rag.NewBatchedEmbedder（ark 1024维）→ milvus2.Indexer.Store
  → governance.audit（document_upload 事件）
```

**检索（Retrieve）**：
```
query + kbID → governance.active strategy_profile（topk/门控阈值）
  → retrieval.DecideDynamicTopK → CandidateTopK
  → retrieval.Searcher（适配现有 HybridRetrieve：Milvus dense+sparse+RRF k=60，topN=CandidateTopK）
  → retrieval.postprocess：去重 → Jaccard 重排 → (引文一致性) → 战略 TopK/分数悬崖/token 预算截断 → 证据门控
  → 返回 Result{Items, Metrics, EvidenceGate}
  → governance.audit（retrieve 事件，可选）
```

---

## 2. 文档标准化（canonical）

### 2.1 契约类型

`canonical/document.go`：
```go
type NormalizedDocument struct {
    ContentMarkdown     string                 // 规范化后 markdown（切块输入）
    ContentMarkdownRaw  string                 // 原始（审计可比对）
    Source              Source                 // FileName/FileType/SourcePath/PageCount
    Blocks             []NormalizedBlock       // 带偏移的块（provenance 用）
    Tables              []NormalizedTable      // 表格 span
    Quality            ParseQuality
    Canonicalization   CanonicalizationInfo   // Version/AppliedRules/RawSHA1/CanonicalSHA1
}
type NormalizedBlock struct{ ID, Type string; Page int; Text string; MarkdownStart, MarkdownEnd int }
type NormalizedTable struct{ ID string; Page int; MarkdownStart, MarkdownEnd int; Rows []TableRow }
type CanonicalizationInfo struct{ Version string; AppliedRules []string; RawSHA1, CanonicalSHA1 string }
```
版本常量 `canonical-normalizer-v1`（对齐源项目）。

### 2.2 Normalize 规则链

`canonical/normalize.go::Normalize(doc)` 按固定顺序在 `ContentMarkdownRaw`（回退 `ContentMarkdown`）上执行，每步记录到 `AppliedRules`：

1. `normalizeUnicodeAndLineEndings` — `\r\n`/`\r` → `\n`，NFKC（`golang.org/x/text/unicode/norm`，awesomeProject4 已通过 eino 间接依赖）。
2. `normalizeLineNoise` — 去解析器噪声（`## ·`、`- ·`、`- -` 等伪项目符号）。
3. `normalizeHeadings` — ATX 标题语法修正（`#{1,6}\s+`）、去强调标记、数字前缀与文字间补空格、CJK 内部去空格。
4. `normalizeHeadingContinuations` — 跨行的标题碎片合并（如 `1.1` + `TERM`）。
5. 折叠 3+ 空行为 2。
6. `normalizeTablesAndSpans` — 重跑 `NormalizeMarkdownPipeTables`，重算表格 MarkdownStart/End。

输出写回 `ContentMarkdown`，保留 `ContentMarkdownRaw`，填 `CanonicalizationInfo`。

### 2.3 表格标准化

`canonical/table.go::NormalizeMarkdownPipeTables`：扫描 pipe-table 行 + 分隔行，解析单元格，渲染规范 pipe table，记录每表 `MarkdownStart/MarkdownEnd`。面试知识库常有表格（问题清单、评分表），此能力高价值。

### 2.4 Provenance

`canonical/provenance.go::AnnotateChunksWithProvenance`：把切块后的 chunk 偏移映射回 `block_ids` / `page`，写入 chunk metadata，支撑「这块来自哪页哪段」。

### 2.5 格式范围

- 本地 txt/md/markdown：规则链全跑。
- PDF：现有 eino pdf parser 产出文本 → 走同一规则链（不接 docling/parser-provider，避免引入外部服务）。
- DOCX：现有 eino docx loader 产出 → 规则链；表格若 eino 未还原为 markdown pipe table，`NormalizeMarkdownPipeTables` 仍尽力识别。
- 不导 HTTP provider / docling。

### 2.6 验收

上传含表格的面试知识 docx/md → 入库 chunk metadata 带 `raw_sha1` / `canonical_sha1` / `applied_rules`；规范化前后对比日志可见应用了哪些规则；表格独立成结构化 span。

---

## 3. 切块策略（chunking）

### 3.1 接口与路由

`chunking/strategy.go`：
```go
type Strategy interface {
    Split(ctx context.Context, req Request) ([]*schema.Document, error)
}
type Request struct {
    Document       *canonical.NormalizedDocument
    BaseMeta       map[string]any
    NormalizedPath string
}
```
常量：`StrategyMarkdown`、`StrategyTableAware`（结构/OCR 感知本轮默认关闭，留接口）。

`chunking/router.go::StrategyRouter`：持有 `defaultStrategy` 与有序 `[]RoutedStrategy{Name, Match, Strategy}`。`Split` 按序匹配，首个命中者切，chunk metadata 记 `chunking_route`。默认装配：table-aware（`MatchHasTables: len(doc.Tables)>0`）→ markdown 兜底。

### 3.2 MarkdownStrategy

`chunking/markdown.go`：用 **goldmark** AST 遍历标题树，按章节切分；每块带 `hierarchy_path` / `section_title` / `heading_level`；字节限长 `MaxChunkBytes=1000`（存储字段上限 4096，留余量）。与现有 `rag.NewRecursiveSplit`（ChunkSize384/Overlap64）的纯文本切法互补——markdown 策略优先按结构，超长段再递归。

### 3.3 TableAwareStrategy

`chunking/table_aware.go`：`doc.Tables` 中每表独立成 chunk，"列名:值"行式渲染，`chunking_unit=table`，保持行完整。

### 3.4 Parent-Child 元数据

`chunking/parent_child.go::finalizeChunks`：为每 chunk 算
- `child_start_offset`/`child_end_offset`（从 metadata 或在 canonical markdown 中定位 chunk 内容）
- `chunk_id`/`child_id`（`document_id`+index+sha1 短哈希）
- `parent_id`、`parent_start_offset`/`parent_end_offset`、`parent_build_strategy`（`heading_section`/`table`/`paragraph_window`）、`parent_token_count`、`section_title`、`hierarchy_path`

父子解析：找包含 child 的最窄 markdown 标题章节为 parent；`chunking_unit=table` 时 parent=表格 span。支撑「小块检索、大上下文生成」（本包先**写入元数据**，父子填充检索留接口、默认关闭，快速验证不阻塞）。

### 3.5 语义二次切分

`chunking/semantic.go`：对超长 chunk（≥ `MinBlockSize` 默认 1200 rune）按句（CJK/ASCII 句末符号）切，复用现有 ark embedder 算相邻句 cosine，按 `BreakpointPercentile=20` 选断点重切，尊重 `TargetChunkSize`/`MaxChunkSize`/`MinSentencesPerChunk`。`semantic_split_enabled` / `semantic_split_score` 写 metadata。默认**开关关闭**（面试文档小、省 embedding 成本），CLI 可开。

### 3.6 默认参数

| 参数 | 默认 | 说明 |
|---|---|---|
| MaxChunkBytes | 1000 | markdown 策略块上限 |
| 重叠 | 64 | 递归切分重叠（沿用现有） |
| SemanticSplit | 关 | 可 CLI 开 |
| BreakpointPercentile | 20 | 语义断点分位 |
| MinBlockSize | 1200 rune | 触发语义切分阈值 |

### 3.7 验收

同一文档走 table-aware 与 markdown 分别产生正确结构；每 chunk 有 `parent_id` 链；评测样例可标注「该 chunk 属于哪个章节」。

---

## 4. 检索增强（retrieval）

### 4.1 定位

**不重做向量检索**。现有 Milvus 混合检索（`rag.HybridRetrieve`：dense COSINE + sparse BM25 + RRF k=60，topK=5）已足够。本包提供**后处理层** `retrieval.postprocess`，叠加在其结果之上，可开关。

### 4.2 Searcher 接口

`retrieval/searcher.go`：
```go
type Searcher interface {
    Search(ctx context.Context, query string, topK int, filter string) ([]*schema.Document, error)
}
```
默认实现 `milvusSearcher` 适配现有 `rag.HybridRetrieve`（topN 取动态 TopK 上限）。可注入 mock 做纯单测。

### 4.3 后处理流水线

`retrieval/postprocess.go` 编排顺序：

**A. 预搜索决策**（在调用 Searcher 前）：
- `DecideDynamicTopK`（`topk.go`）：由查询长度/词数/宽泛度算 `CandidateTopK`（搜索召回上限）与 `FinalTopK`（最终返回数）。策略版本 `phase2-rule-v1`。Searcher 以 `CandidateTopK` 为 topN 搜索。

**B. 后处理**（`PostProcess(ctx, query, hits, profile) (Result, error)`，对搜索结果顺序执行）：
1. `Dedupe`（`chunking/dedupe.go`，按 `document_id:chunk_id`）。
2. `JaccardReranker.Rerank`（`rerank.go`）：原分 × 词法 Jaccard/覆盖（CJK bigram），`rerank_score`。无外部模型依赖，纯兜底。模型重排留接口、默认关。
3. `CitationConsistencyChecker.Check`（`citation.go`）：从 query 抽 claim，对每 doc 评分，标 `citation_supported`/`citation_support_score`（默认关）。
4. `ApplyScoreCliffGuard` + `DecideStrategicTopK`：按 rerank 分数分布/分数悬崖/证据密度把 `CandidateTopK` 收敛到 `FinalTopK`，超 `TokenBudget` 截断。
5. `EvaluateEvidenceGate`（`evidence_gate.go`）：`pass`/`refused`/`degraded_pass`/`disabled`。5 类拒绝：`No-Retrieval-Hit`/`Low-Rerank-Confidence`/`Insufficient-Citation-Coverage`/`Contradictory-Evidence`/`Out-Of-KB-Scope`。阈值来自 active `strategy_profile`。拒绝时返回空 items + refusal。

返回 `Result{Items, Metrics, EvidenceGate{Outcome, Reason}}`。`Metrics` 记各阶段计数供观测（不写库，留接口）。

### 4.4 默认开关

后处理各步默认**轻量**：去重+Jaccard 重排+证据门控开；动态 TopK 开；引文一致性默认关（面试场景 query 多为短问题，claim 抽取噪声大）。均可经 `strategy_profile` 配置。

### 4.5 检索观测日志

`retrieve_log` 表**不在本包建**（属治理范围，精简版先不做持久化）。`Result.Metrics` 结构化返回，调用方按需写库/打日志。

### 4.6 验收

低相关 query 被门控拒绝带 `refusal_reason`；宽泛 query 动态 TopK 扩大召回；`Result.Metrics` 可追溯决策链。纯单测用 mock Searcher 覆盖各分支。

---

## 5. 治理（governance，精简版）

### 5.1 范围

精简治理 = 知识库元数据 + 索引重建 + 索引健康 + 策略档案（DB 持久化）+ 审计事件。**不做** eval 评测、HTTP 治理路由。

### 5.2 知识库元数据与索引

`governance/knowledge_base.go`：
- `KnowledgeBase` 元数据（id/name/collection/embedding_model/dimension/metric/created_at）——单 collection `documents` 场景下记录配置版本。
- `Reindex(kbID)`：drop & recreate Milvus collection（沿用现有 indexer schema），重跑入库。供 CLI `reindex` 子命令。
- `HealthCheck(kbID)`：collection 存在 + 维度匹配 + 度量类型匹配 + load 健康 + 查询冒烟。返回 `IndexHealth{OK, Checks, Gaps}`。对照源项目「向量契约」的维度/度量健康检查。

### 5.3 策略档案（DB 持久化，修正源项目内存态缺陷）

`governance/strategy_profile.go` + `backend/repository/dao/strategy_profile.go`：
```go
type StrategyProfile struct {
    ID                      uint64 `gorm:"primaryKey"`
    Name                    string `gorm:"uniqueIndex"`
    Status                  string // active / candidate / baseline
    FusionStrategy          string // 留字段（本包不重做融合，默认 milvus_rrf）
    TopKConfig              string // JSON: candidateTopK/finalTopK/动态规则开关
    EvidenceGateThresholds  string // JSON: minRerankScore/minCitationCoverage 等
    CreatedAt, UpdatedAt    time.Time
}
```
- `GetActiveProfile()`：读 status=active（单写者 + 行锁，避免并发）。
- `Activate(id)`：当前 active → baseline，目标 → active，写审计。
- `Rollback(id)`：回滚到 baseline，写审计。
- 检索链路读 active profile 注入 `retrieval.PostProcess`。

### 5.4 审计事件

`governance/audit.go` + `backend/repository/dao/audit_event.go`：
```go
type AuditEvent struct {
    ID                uint64
    AuditTraceID      string `gorm:"uniqueIndex"`
    Operator          string
    Action            string // document_upload / strategy_activate / strategy_rollback / reindex / retrieve
    ResourceType, ResourceID string
    Before, After     string // JSON 快照
    Result, Reason    string
    IP                string
    SensitiveMasked   bool
    CreatedAt         time.Time
}
```
- `persistAuditEvent` 通用写入 + 动作常量表 + 敏感字段脱敏（query/snippet → `[masked]`）。
- 接入点：文档上传、策略激活/回滚、索引重建（检索审计默认关，量大）。

### 5.5 注册

新表 `strategy_profile`、`audit_event`、（可选）`knowledge_base` 加入 `Init/mysql.go::AutoMigrate`。各自 `dao`+`repository`+Wire `providerSet` 条目（沿用现有 DDD 分层）。**不新增 HTTP 路由**；治理操作经 CLI（`ragkit-cli reindex/activate/rollback`）触发。

### 5.6 验收

切换策略版本 → 检索行为变化 → 审计事件可查；`HealthCheck` 能发现维度/度量不一致；`Reindex` 能重建集合。无 HTTP 路由也能完成治理闭环（CLI 驱动）。

---

## 6. 接线点（仅 stub，零逻辑改动）

### 6.1 入库接线 `agent/knowledge/import.go`

在现有 `loader → splitter` 之间，以**注释占位 + 可选开关**插入：
```go
docs, _ := rag.NewDocumentsLoader(ctx, path)
// === ragkit 接线点（标准化 + 路由切块），默认关闭，手动切换时启用 ===
// if os.Getenv("RAGKIT_ENABLED") == "1" {
//     nd := ragkit.NormalizeDocs(docs)        // canonical
//     docs = ragkit.Split(ctx, nd, ragkit.DefaultRouter()) // chunking
// }
splitter := rag.NewSemanticSplit(...)  // 现有逻辑不变
```
切换时：用户把注释内代码启用、或直接调用 `ragkit` facade。**不删现有代码**，保底可回退。

### 6.2 检索接线 `agent/tool/retriever.go`

`GetRetrieverWithInput` 内同样以 stub 包装：
```go
// === ragkit 接线点（门控 + 动态 TopK），默认关闭 ===
// if os.Getenv("RAGKIT_ENABLED") == "1" {
//     return ragkit.Retrieve(ctx, query, topK, filter)
// }
return rag.HybridRetrieve(ctx, client, query, topK, filter)  // 现有不变
```

### 6.3 不碰清单

`agent/agents/*`、`agent/service/*`、面试 SSE 链路 `backend/api/handler/interview_run/*`、`backend/api/handler/handler.go::Register` —— 全部零改动。切换不影响面试对话，可独立回退。

---

## 7. 验证策略

### 7.1 独立验证 CLI `agent/knowledge/ragkit/cmd/ragkit-cli`

子命令：
- `ragkit-cli ingest <path>`：loader → canonical.Normalize → StrategyRouter.Split → 现有 embedder+indexer 入库，打印 chunk 数 / 双 SHA1 / applied_rules / 路由标签。
- `ragkit-cli retrieve <query>`：Searcher → postprocess → 打印 items + Metrics + EvidenceGate。
- `ragkit-cli reindex`：drop & recreate collection，重跑入库。
- `ragkit-cli health`：IndexHealth 检查。
- `ragkit-cli activate <profile>` / `rollback <profile>`：策略切换 + 审计。

CLI 直接读 `.env`（沿用 `godotenv.Load(".env")`），复用 `Init.NewMilvusManger`，无需起 HTTP 服务即可全链路验证。

### 7.2 单元测试

- `canonical`：规则链每步表驱动测试（含 CJK、表格、标题续行）。
- `chunking`：mock `NormalizedDocument`，验证 markdown/table-aware 切块结构 + parent-child 元数据；语义切分用 mock embedder。
- `retrieval`：mock `Searcher`，覆盖去重/TopK/Jaccard/引文/门控各分支。
- `governance`：strategy_profile 并发读写用 `sqlmock` 或内存 sqlite；audit 脱敏用例。
- 复用现有 `retrieval_eval_test.go` 风格（env gate + 纯函数）。

### 7.3 E2E 验收

`doc/*.docx`（含表格）→ `ragkit-cli ingest` → `ragkit-cli retrieve "面试问题"` → 命中带 provenance/parent_id；构造低相关 query → 门控拒绝带 reason；`ragkit-cli activate` 切 profile → 检索行为变化 → `audit_event` 可查。

### 7.4 快速验证标准

- `go build ./agent/knowledge/ragkit/...` 通过，**不破坏** `go build ./...`（现有代码零改动）。
- 现有 `agent/knowledge/rag` 包测试仍通过（未被修改）。
- CLI 三个核心子命令（ingest/retrieve/health）在本地 `.env` 下跑通。

---

## 8. 依赖与风险

- **新增依赖**：`github.com/yuin/goldmark`（Markdown AST 切块，源项目同款，与 eino 兼容，风险低）。
- **复用**：`golang.org/x/text/unicode/norm`（eino 间接已有）、eino milvus2、ark embedder、GORM、godotenv。
- **风险**：
  1. 现有 `.env` 含真实密钥（Zilliz/aihubmix）——审计脱敏要覆盖 query/snippet，但密钥本身不在审计范围。
  2. 策略 profile 并发读写——单写者 + 行锁，CLI 单线程触发为主。
  3. 语义切分/索引重建调用 embedding——面试文档量小（<100），成本可控，默认关闭。
  4. **不碰现有面试 agent**——保底可随时单独回退到现有 `rag` 包。

---

## 9. 不在本设计范围（明确排除）

- eval 评测 runner / eval 数据集 / eval 用例 / eval run 表与 Kafka 异步执行。
- `/eval/*` `/strategy/*` `/audit/*` HTTP 治理路由（治理经 CLI）。
- 在 Milvus 之上再叠 minmax/RRF 融合层（现有 RRF k=60 已用）。
- 查询改写（LLM 依赖重，面试场景收益低）。
- 语义缓存、A/B 实验、租户体系、索引 registry 多 collection 切换。
- agentic shadow 切块、HTTP parser-provider/docling。
- DOCX/PDF 解析器重写（沿用 eino loader）。
- 父子填充检索（先写元数据，填充留接口默认关）。

后续如需 eval 评测或 HTTP 治理路由，可在本包基础上按 `doc/rag-governance-import-plan.md` Phase 4 扩展，本设计已为其留好 `Searcher` 接口与 `strategy_profile` 表的扩展点。
