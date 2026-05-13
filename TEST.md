# document TEST

## 覆盖场景
- [x] **Excel 基础读写**: 验证 `NewExcel`, `Set`, `Get`, `Save`, `OpenExcel`。
- [x] **Excel 对象映射**: 验证 `SetData`, `GetData` 及动态列扩展。
- [x] **统一 API**: 验证 `document.Open` 自动识别、`ToJSON` 和 `ToMarkdown`。
- [x] **ID 生成解析**: 验证 `MakeCellID` 和 `ParseCellID` 的准确性。
- [x] **Docx 标题识别**: 通过 XML 提取带 `#` 的层级 Markdown。
- [x] **Graph 关系文档**: 验证 `Node` 增删、`ToMarkdown` (Mermaid) 及 `Save/Open`。
- [x] **CSV/Markdown 支持**: 验证 CSV 转对象/表格，Markdown 文件的统一接口读写。
- [x] **多 Sheet Excel**: 验证 `ToMarkdown` 导出所有工作表。
- [x] **PDF 页码标记**: 验证提取内容中包含页码注释。

## 性能测试 (Benchmark)
- **环境**: Darwin / Apple M3 Max
- **Excel.SetData**: 写入 100 行对象数据。

```text
BenchmarkExcel_SetData-8             712           1673885 ns/op
```

## 基础设施对齐
- 全面使用 `apigo.cc/go/cast` 替代原生类型转换。
- 移除了所有原生 `os` 和 `strconv` 依赖。
- Docx 提取原生集成，不再强制依赖第三方 binary 提取工具。
