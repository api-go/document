# CHANGELOG

## v1.0.5 (2026-05-13)
- **新特性**: 引入 `Graph` 关系型文档类型 (`.graph`)，支持节点内容与关系的统一建模。
- **新特性**: 关系型文档支持一键转换为带 **Mermaid** 图表的 Markdown，适用于脑图与分镜。
- **功能增强**: 改进 `Docx` 提取算法，支持自动识别标题层级并转换为标准的层级 Markdown (H1-H6)。
- **架构对齐**: `OpenGraph` 支持文件不存在时自动创建，与 `OpenExcel` 行为一致。
- **文档更新**: README 增加 RAG 增强与关系文档示例。

## v1.0.4 (2026-05-12)
- **依赖同步**：同步更新 `go/cast` 和 `go/file` 至最新版本。
- **文档优化**：完善 README 与测试用例。

## v1.0.3 (2026-05-12)

- 重构并更名为 `document`，统一处理 Excel, Word, PPT, PDF。
- 引入 `Document` 接口，支持 **Open/Create/Save** 的极简 API。
- 所有文档支持载体化操作：`ToJSON()` 和 `ToMarkdown()`。
- 基础设施全面对齐 `@go` 标准库。
- 修复 Excel 动态列扩展与 JSON 互转逻辑。
