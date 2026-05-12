# office

极简、高效的 Go Excel 处理库，基于 `excelize` 构建，符合 `@go` 设计哲学。

## 特性

- **统一 API**: 提供极简的 `Open`, `Save`, `Get`, `Set` 等操作。
- **自动对齐**: 自动处理工作表的创建和索引。
- **对象映射**: 支持将 `[]map[string]any` 直接写入 Excel 或从 Excel 读取。
- **高性能**: 尽量减少内存分配和冗余操作。

## 快速开始

### 安装

```bash
go get apigo.cc/go/office
```

### 基础用法

```go
import "apigo.cc/go/office"

// 创建并写入
xls := office.New()
table := [][]any{
    {"Name", "Age"},
    {"Alice", 25},
    {"Bob", 30},
}
xls.Set("Sheet1", table, "A1", "")
xls.Save("example.xlsx")

// 读取
xls2, _ := office.Open("example.xlsx")
data, _ := xls2.Get("Sheet1", "A1", "")
```

### 对象列表操作

```go
data := []map[string]any{
    {"Name": "Alice", "Age": 25},
    {"Name": "Bob", "Age": 30},
}
xls.SetData("Users", data, "A1", "")
```

## API 参考

### 核心函数

- `New() *Excel`: 创建新的 Excel 对象。
- `Open(filename string, password ...string) (*Excel, error)`: 打开现有文件。

### Excel 方法

- `Save(filename ...string) error`: 保存文件。
- `Bytes() ([]byte, error)`: 获取字节切片。
- `Set(sheetName string, table [][]any, start, end string) error`: 写入二维数据。
- `Get(sheetName string, start, end string) ([][]any, error)`: 读取二维数据。
- `SetData(sheetName string, data []map[string]any, start, end string) error`: 写入对象列表。
- `GetData(sheetName string, start, end string) ([]map[string]any, error)`: 读取对象列表。
- `Sheets() []string`: 获取工作表列表。
- `RemoveSheet(sheetName string) error`: 删除工作表。

### 工具函数

- `MakeCellID(col, row int) string`: 生成单元格 ID (如 "A1")。
- `ParseCellID(cell string) (col, row int)`: 解析单元格 ID。
- `MakeColID(col int) string`: 生成列 ID (如 "A")。
