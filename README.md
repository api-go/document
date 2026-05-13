# document

极简、高效的 Go 办公文档处理桥梁，符合 `@go` 设计哲学。

## 核心设计

`document` 包将所有复杂的文档格式（Excel, Word, PPT, PDF, Graph）抽象为统一的 `Document` 接口。你只需要关心 **数据载体** (JSON/Markdown)，而不需要关心底层实现。

- **结构化数据 (Excel)** <-> **JSON**
- **半结构化内容 (Word/PPT/PDF)** -> **结构化 Markdown** (支持标题识别)
- **关系型内容 (Graph/Mindmap)** <-> **JSON/Mermaid**

## 统一 API

```go
import "apigo.cc/go/document"

// 1. 打开文档 (自动识别类型)
doc, _ := document.Open("data.xlsx")

// 2. 转换为通用载体
jsonStr := doc.ToJSON()    // 适合 Excel / Graph
mdStr := doc.ToMarkdown() // 适合 Word / PDF / Graph (Mermaid)

// 3. 修改并保存 (Excel/Graph 支持数据回写)
if xls, ok := doc.(*document.Excel); ok {
    xls.SetData("Sheet1", newData, "A1", "")
}
doc.Save("updated.xlsx")
```

## 支持格式

| 格式 | 后缀 | `ToJSON` 表现 | `ToMarkdown` 表现 | `Save` 行为 |
| :--- | :--- | :--- | :--- | :--- |
| **Excel** | `.xlsx` | 返回第一个 Sheet 的对象数组 | 返回**所有 Sheet** 的 MD 表格 | 保存为 Excel |
| **Word** | `.docx` | 返回 `{"metadata":..., "content":...}` | 返回**带层级标题**的 Markdown | 保存为文本 |
| **PPT** | `.pptx` | 返回 `{"metadata":..., "content":...}` | 返回提取的全文文本 | 保存为文本 |
| **PDF** | `.pdf` | 返回 `{"metadata":..., "content":...}` | 返回**带页码标记**的 Markdown | 保存为文本 |
| **Graph** | `.graph` | 返回节点与关系的结构化 JSON | 返回 **Mermaid 关系图** + 节点详情 | 保存为 JSON |
| **CSV** | `.csv` | 返回对象数组 | 返回 MD 表格 | 保存为 CSV |
| **Markdown**| `.md` | 返回 `{"content":...}` | 返回原始文本 | 保存为 Markdown |
| **Text** | `.txt` | 返回 `{"content":...}` | 返回原始文本 | 保存为文本 |

## 快速示例

### AI 友好型转换 (RAG 增强)

```go
// 一行代码将 Word 转为带层级标题的 Markdown，方便 AI 分片 (Chunking)
md, _ := document.Open("manual.docx").ToMarkdown()
```

### 关系型文档 (脑图/小说大纲/分镜脚本)

```go
g := document.NewGraph()
g.Title = "故事大纲"
g.AddNode(&document.Node{ID: "1", Title: "开端", Content: "主角获得系统", Type: "plot"})
g.AddNode(&document.Node{ID: "2", Title: "冲突", Content: "反派出现", Links: []string{"1"}})
mdStr := g.ToMarkdown() // 包含 Mermaid graph TD 的渲染内容
```

## API 参考

### 顶级函数
- `Open(filename string, password ...string) (Document, error)`
- `Create(ext string) (Document, error)`

### Document 接口
- `ToJSON() string`
- `ToMarkdown() string`
- `Save(filename ...string) error`

### Graph 专用 (关系型文档)
- `AddNode(n *Node)`
- `OpenGraph(filename string) (*Graph, error)`
