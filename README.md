# office

极简、高效的 Go Office 文档处理库，符合 `@go` 设计哲学。支持将 Office 文档与 JSON/Markdown 等通用格式无缝转换，作为 AI 时代的数据桥梁。

## 特性

- **统一 API**: 提供极简的 `Open`, `Save`, `ToJSON`, `ToMarkdown` 等操作。
- **载体转换**: 
  - **Excel <-> JSON**: 自动将表格映射为对象数组，支持双向转换。
  - **Docx/Pptx/PDF -> Markdown**: 提取并转化为 AI 友好的 Markdown 格式。
- **纯 Go 实现**: 无 CGo 依赖，极致的跨平台与部署性能。
- **Excel 增强**: 自动处理工作表对齐，支持动态列扩展。

## 快速开始

### 安装

```bash
go get apigo.cc/go/office
```

### Excel <-> JSON (结构化数据载体)

```go
// Excel 转 JSON
xls, _ := office.Open("data.xlsx")
jsonStr, _ := xls.ToJSON("Sheet1", "A1", "")

// JSON 转 Excel
newXls := office.New()
newXls.FromJSON("Sheet1", jsonStr, "A1", "")
newXls.Save("restored.xlsx")
```

### Docx/Pptx/PDF -> Markdown (内容载体)

```go
// Word 转 Markdown
doc, _ := office.OpenDocx("contract.docx")
md, _ := doc.ToMarkdown()

// PDF 转 Markdown
pdf, _ := office.OpenPDF("report.pdf")
md, _ := pdf.ToMarkdown()
```

## API 参考

### Excel (JSON 转换)
- `ToJSON(sheetName string, start, end string) (string, error)`
- `FromJSON(sheetName string, jsonStr string, start, end string) error`
- `GetData(sheetName string, start, end string) ([]map[string]any, error)`
- `SetData(sheetName string, data []map[string]any, start, end string) error`

### Word/PPT/PDF (Markdown 提取)
- `ToMarkdown() (string, error)`: 将文档内容提取为 Markdown 格式字符串。
- `Text() (string, error)`: 提取纯文本。
