# CHANGELOG

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
