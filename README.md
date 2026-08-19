# 面试舱 · AI 面试训练台

基于 eino 多 Agent 编排 + RAG 知识库的 AI 模拟面试系统：上传简历 → 智能解析生成画像 → 多场景面试 → 流式作答 → 异步生成评估报告与 Offer 预测。

## 功能

- **简历管理**：简历上传与解析
- **模拟面试**：AI 多轮问答、流式作答
- **面试评估**：面试结束后经 Kafka 异步触发，产出全面评估报告、分维度评估详情与 Offer 预测
- **RAG 知识库（自研 ragkit）**：简历 / 知识文档分块（语义、表格感知、父子分块）、混合检索、重排、引用溯源
- **用户体系**：JWT 注册 / 登录鉴权

## 技术栈

Go 1.25 · Gin · GORM (MySQL) · Redis · Kafka · Milvus · eino · Wire · zap

## 快速开始

```bash
# 需 MySQL / Redis / Kafka / Milvus，连接配置见 backend/Init/ 与 .env
go mod download && go run main.go      # :8080
```

## 目录结构

```
main.go            入口（服务 + Kafka 评估消费组）
agent/             AI 编排层：面试官 / 简历解析 / 评估 / 预测 Agent，ragkit RAG 知识库
backend/           Init 初始化、api 路由与 handler、event（Kafka）、repository（GORM）
```

## API 概览

| 分组 | 接口 | 说明 |
| --- | --- | --- |
| `/user` | POST /register、/login | 注册登录 |
| `/resume` | GET /list、POST /upload、DELETE /:id、PUT /:id/default | 简历管理 |
| `/mianshi` | POST /stream/start、/stream/resume、/answer/submit、/interview/end、GET /session/info | 面试流程 |
| `/evaluation` | POST /report、/predict | 评估报告、Offer 预测 |
