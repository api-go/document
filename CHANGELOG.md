# CHANGELOG

## v1.0.4 (2026-05-12)
- **依赖同步**：同步更新 `go/cast` 和 `go/file` 至最新版本。
- **文档优化**：完善 README 与测试用例。

## v1.0.3 (2026-05-12)

- 重构并更名为 `document`，统一处理 Excel, Word, PPT, PDF。
- 引入 `Document` 接口，支持 **Open/Create/Save** 的极简 API。
- 所有文档支持载体化操作：`ToJSON()` 和 `ToMarkdown()`。
- 基础设施全面对齐 `@go` 标准库。
- 修复 Excel 动态列扩展与 JSON 互转逻辑。
