# office

极简、高效的 Go Office 文档处理桥梁，符合 `@go` 设计哲学。

## 核心设计

`office` 包将所有复杂的文档格式（Excel, Word, PPT, PDF）抽象为统一的 `Document` 接口。你只需要关心 **数据载体** (JSON/Markdown)，而不需要关心底层实现。

- **结构化数据 (Excel)** <-> **JSON**
- **半结构化内容 (Word/PPT/PDF)** -> **Markdown**

## 统一 API

```go
import "apigo.cc/go/office"

// 1. 打开文档 (自动识别类型)
doc, _ := office.Open("data.xlsx")

// 2. 转换为通用载体
jsonStr := doc.ToJSON()    // 适合 Excel
mdStr := doc.ToMarkdown() // 适合 Word/PPT/PDF

// 3. 修改并保存 (Excel 支持数据回写)
// (对于 Excel，你可以强制断言获取更多方法)
if xls, ok := doc.(*office.Excel); ok {
    xls.SetData("Sheet1", newData, "A1", "")
}
doc.Save("updated.xlsx")
```

## 支持格式

| 格式 | 后缀 | `ToJSON` 表现 | `ToMarkdown` 表现 | `Save` 行为 |
| :--- | :--- | :--- | :--- | :--- |
| **Excel** | `.xlsx` | 返回第一个 Sheet 的对象数组 | 返回第一个 Sheet 的 MD 表格 | 保存为 Excel 文件 |
| **Word** | `.docx` | 返回 `{"metadata":..., "content":...}` | 返回提取的全文文本 | 保存为纯文本文件 |
| **PPT** | `.pptx` | 返回 `{"metadata":..., "content":...}` | 返回提取的全文文本 | 保存为纯文本文件 |
| **PDF** | `.pdf` | 返回 `{"metadata":..., "content":...}` | 返回提取的全文文本 | 保存为纯文本文件 |

## 快速示例

### AI 友好型转换

```go
// 一行代码将 PDF 转为 AI 易读的 Markdown
md, _ := office.Open("report.pdf").ToMarkdown()

// 一行代码将 Excel 转为结构化 JSON
json, _ := office.Open("sales.xlsx").ToJSON()
```

### 极简读写 (Excel)

```go
xls, _ := office.OpenExcel("config.xlsx")
data, _ := xls.GetData("Sheet1", "A1", "")
// ... 修改 data
xls.SetData("Sheet1", data, "A1", "")
xls.Save()
```

## API 参考

### 顶级函数
- `Open(filename string, password ...string) (Document, error)`
- `Create(ext string) (Document, error)`

### Document 接口
- `ToJSON() string`
- `ToMarkdown() string`
- `Save(filename ...string) error`

### 类型特定 (导出)
- `Excel`: 支持 `SetData`, `GetData`, `Sheets` 等。
- `Docx`, `Pptx`, `PDF`: 支持 `Content` 和 `Metadata` 直接访问。
