# Test Report: @go/document

## 📋 测试概览
- **测试时间**: 2026-06-21
- **测试环境**: darwin/amd64
- **Go 版本**: 1.25.0

## ✅ 功能测试 (Functional Tests)
| 场景 | 状态 | 描述 |
| :--- | :--- | :--- |
| `TestExcel_Basic` | PASS | Excel 基础单元格及公式读写测试。 |
| `TestExcel_Data` | PASS | Excel 对象数组及映射读写测试。 |
| `TestUnifiedAPI` | PASS | 统一的 Document 接口 Open/Create/Save 行为测试。 |
| `TestIDGeneration` | PASS | RAG 场景 ID 生成器及映射验证。 |
| `TestGraph_Basic` | PASS | Graph 关系型文档节点及关系建模测试。 |
| `TestGraph_Unified` | PASS | Graph 到 Markdown (Mermaid) 链路集成测试。 |
| `TestCSV` | PASS | CSV 文档对象及表格转换测试。 |
| `TestMarkdown` | PASS | Markdown 格式解析与重构测试。 |

## 🛡️ 鲁棒性防御 (Robustness)
- **可选参数提示**：`Open` 和 `Save` 方法接受指针作为可选参数，结合最新的 `go/js` 引擎智能过滤空参数。
- **JS 错误调用栈**：JS 桥接层改用具名导出并使用 `jsmod.MakeError` 包裹错误，确保 JS 抛出异常时携带准确的 Go 运行时堆栈。

## ⚡ 性能基准 (Benchmarks)
| 函数 | 平均耗时 | 性能分析 |
| :--- | :--- | :--- |
| `Excel_SetData` | **1874250 ns/op** | Excel 大量数据处理性能指标。 |
