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
