# CHANGELOG

## v1.3.0 (2026-05-12)

- 核心架构重构：引入统一的 `Document` 接口，支持 **Open/Create/Save** 的极简 API 范式。
- 自动识别：`office.Open(path)` 可根据后缀名自动选择解析器。
- 标准化载体：所有文档类型均支持 `ToJSON()` 和 `ToMarkdown()`，完美对齐 AI 工作流。
- 零摩擦设计：Excel 导出自动处理 Sheet 和列扩展；PDF/Word/PPT 提取自动结构化。

## v1.2.0 (2026-05-12)

- 新增 Excel 与 JSON 的双向转换支持 (`ToJSON`, `FromJSON`)。
- 为 Word、PPT 和 PDF 统一增加 `ToMarkdown` 方法，提升 AI 友好度。
- 强化 Excel 导出逻辑，支持从 JSON 直接生成结构化表格。

## v1.1.0 (2026-05-12)

- 新增 Word (`.docx`) 解析支持，可提取全文纯文本。
- 新增 PowerPoint (`.pptx`) 解析支持，可提取幻灯片纯文本。
- 新增 PDF 解析支持，可提取全文纯文本及元数据（页数、作者等）。
- 保持全纯 Go 实现，无 CGo 依赖。

## v1.0.0 (2026-05-12)

- 从 `gojs/office` 迁移并重构为纯 Go 实现。
- 基础设施对齐：使用 `apigo.cc/go/cast` 和 `apigo.cc/go/file`。
- 移除 `gojs` 和 `goja` 相关依赖。
- 优化 `SetData` 和 `GetData` 的逻辑，支持自动扩展列。
- 提供更简洁的 `Open` 和 `New` 入口。
