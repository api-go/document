# office

极简、高效的 Go Office 文档处理库，符合 `@go` 设计哲学。支持 Excel、Word (Docx)、PowerPoint (Pptx) 和 PDF 的解析与处理。

## 特性

- **统一 API**: 提供极简的 `Open`, `Save`, `Text` 等操作。
- **纯 Go 实现**: 无 CGo 依赖，跨平台支持。
- **解析与识别**: 支持从 Docx、Pptx 和 PDF 中提取纯文本内容。
- **Excel 增强**: 自动处理工作表对齐，支持对象列表 (`[]map`) 的直接读写。

## 快速开始

### 安装

```bash
go get apigo.cc/go/office
```

### Excel 处理

```go
// 写入数据
xls := office.New()
xls.Set("Sheet1", [][]any{{"Name", "Age"}, {"Alice", 25}}, "A1", "")
xls.Save("example.xlsx")

// 读取对象列表
xls2, _ := office.Open("example.xlsx")
data, _ := xls2.GetData("Sheet1", "A1", "")
```

### Word (Docx) 解析

```go
doc, _ := office.OpenDocx("contract.docx")
text, _ := doc.Text() // 提取全文文本
fmt.Println(text)
```

### PowerPoint (Pptx) 解析

```go
ppt, _ := office.OpenPptx("presentation.pptx")
text, _ := ppt.Text() // 提取幻灯片全文
```

### PDF 解析

```go
pdf, _ := office.OpenPDF("report.pdf")
text, _ := pdf.Text() // 提取 PDF 纯文本
info := pdf.Info()   // 获取页数、作者等元数据
```

## API 参考

### Excel
- `New() *Excel`
- `Open(filename string, password ...string) (*Excel, error)`
- `Set(sheetName string, table [][]any, start, end string) error`
- `SetData(sheetName string, data []map[string]any, start, end string) error`

### Word (Docx)
- `OpenDocx(filename string) (*Docx, error)`
- `Text() (string, error)`

### PowerPoint (Pptx)
- `OpenPptx(filename string) (*Pptx, error)`
- `Text() (string, error)`

### PDF
- `OpenPDF(filename string) (*PDF, error)`
- `Text() (string, error)`
- `Info() map[string]any`
