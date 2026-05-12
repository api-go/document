package office

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"

	"apigo.cc/go/cast"
	"apigo.cc/go/file"
	"github.com/xuri/excelize/v2"
)

// Excel 封装了 Excel 文件的核心操作，提供极简且高效的 API。
type Excel struct {
	filename string
	password string
	excel    *excelize.File
}

// New 创建一个新的 Excel 对象。
func New() *Excel {
	return &Excel{
		excel: excelize.NewFile(),
	}
}

// Open 打开一个现有的 Excel 文件。
// 如果文件不存在，将创建一个新的 Excel 对象。
func Open(filename string, password ...string) (*Excel, error) {
	pwd := ""
	if len(password) > 0 {
		pwd = password[0]
	}

	xls := &Excel{
		filename: filename,
		password: pwd,
	}

	if file.Exists(filename) {
		f, err := excelize.OpenFile(filename, excelize.Options{Password: pwd})
		if err != nil {
			return nil, err
		}
		xls.excel = f
	} else {
		xls.excel = excelize.NewFile()
	}

	return xls, nil
}

// Save 保存 Excel 文件。
// 如果提供了 filename，将另存为该文件名。
func (xls *Excel) Save(filename ...string) error {
	if len(filename) > 0 && filename[0] != "" {
		xls.filename = filename[0]
	}
	if xls.filename == "" {
		return fmt.Errorf("no filename specified")
	}
	return xls.excel.SaveAs(xls.filename, excelize.Options{Password: xls.password})
}

// Bytes 将 Excel 内容写入字节切片。
func (xls *Excel) Bytes() ([]byte, error) {
	buf, err := xls.excel.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Set 设置指定单元格范围的值。
// table: 二维数组，代表行和列。
// start: 起始单元格 ID（如 "A1"），默认为 "A1"。
// end: 结束单元格 ID（如 "C3"），用于限制写入范围。
func (xls *Excel) Set(sheetName string, table [][]any, start, end string) error {
	sheet := xls.getOrCreateSheet(sheetName)
	startX, startY := ParseCellID(start)
	endX, endY := ParseCellID(end)

	for y, row := range table {
		if endY > 0 && startY+y > endY {
			break
		}
		for x, v := range row {
			if endX > 0 && startX+x > endX {
				break
			}
			cellID := MakeCellID(startX+x, startY+y)
			if err := xls.excel.SetCellValue(sheet, cellID, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// Get 获取指定范围的单元格数据。
func (xls *Excel) Get(sheetName string, start, end string) ([][]any, error) {
	sheet := xls.getOrCreateSheet(sheetName)
	rows, err := xls.excel.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	startX, startY := ParseCellID(start)
	endX, endY := ParseCellID(end)

	result := make([][]any, 0)
	for y, row := range rows {
		if startY > 0 && y < startY {
			continue
		}
		if endY > 0 && y > endY {
			break
		}

		rowData := make([]any, 0)
		for x, v := range row {
			if startX > 0 && x < startX {
				continue
			}
			if endX > 0 && x > endX {
				break
			}

			// 尝试根据单元格类型转换数据
			cellID := MakeCellID(x, y)
			cellType, _ := xls.excel.GetCellType(sheet, cellID)
			
			switch cellType {
			case excelize.CellTypeNumber:
				if isFloat(v) {
					rowData = append(rowData, cast.To[float64](v))
				} else {
					rowData = append(rowData, cast.To[int64](v))
				}
			case excelize.CellTypeBool:
				rowData = append(rowData, cast.To[bool](v))
			case excelize.CellTypeDate:
				// TODO: 更好的日期处理
				rowData = append(rowData, v)
			default:
				rowData = append(rowData, v)
			}
		}
		result = append(result, rowData)
	}
	return result, nil
}

// RemoveSheet 删除工作表。
func (xls *Excel) RemoveSheet(sheetName string) error {
	return xls.excel.DeleteSheet(sheetName)
}

// Sheets 返回所有工作表名称。
func (xls *Excel) Sheets() []string {
	return xls.excel.GetSheetList()
}

// SetColWidth 设置列宽。
func (xls *Excel) SetColWidth(sheetName string, startCol, endCol string, width float64) error {
	return xls.excel.SetColWidth(xls.getOrCreateSheet(sheetName), startCol, endCol, width)
}

// SetColWidths 批量设置列宽。
func (xls *Excel) SetColWidths(sheetName string, widths []float64) error {
	sheet := xls.getOrCreateSheet(sheetName)
	for i, w := range widths {
		col := MakeColID(i)
		if err := xls.excel.SetColWidth(sheet, col, col, w); err != nil {
			return err
		}
	}
	return nil
}

// SetCellStyle 设置单元格样式。
func (xls *Excel) SetCellStyle(sheetName string, start, end string, styleID int) error {
	return xls.excel.SetCellStyle(xls.getOrCreateSheet(sheetName), start, end, styleID)
}

// MakeStyle 创建样式并返回样式 ID。
func (xls *Excel) MakeStyle(style *excelize.Style) (int, error) {
	return xls.excel.NewStyle(style)
}

// SetPanes 设置冻结窗格。
func (xls *Excel) SetPanes(sheetName string, panes *excelize.Panes) error {
	return xls.excel.SetPanes(xls.getOrCreateSheet(sheetName), panes)
}

// SetAutoFilter 设置自动筛选。
func (xls *Excel) SetAutoFilter(sheetName string, start, end string, options []excelize.AutoFilterOptions) error {
	return xls.excel.AutoFilter(xls.getOrCreateSheet(sheetName), start+":"+end, options)
}

// SetData 将对象列表写入 Excel。
func (xls *Excel) SetData(sheetName string, data []map[string]any, start, end string) error {
	table, err := xls.Get(sheetName, start, end)
	if err != nil {
		return err
	}

	fieldIndex := map[string]int{}
	fieldNum := 0
	if len(table) > 0 {
		fieldNum = len(table[0])
		for i, v := range table[0] {
			fieldIndex[cast.To[string](v)] = i
		}
	} else {
		table = append(table, []any{})
	}

	for i, item := range data {
		for len(table) <= i+1 {
			table = append(table, make([]any, fieldNum))
		}
		for k, v := range item {
			idx, ok := fieldIndex[k]
			if !ok {
				idx = fieldNum
				fieldIndex[k] = idx
				fieldNum++
				// 扩展所有行以匹配新的列数
				for r := range table {
					for len(table[r]) < fieldNum {
						table[r] = append(table[r], "")
					}
				}
				table[0][idx] = k
			}
			table[i+1][idx] = v
		}
	}
	return xls.Set(sheetName, table, start, end)
}

// GetData 从 Excel 中读取对象列表。
func (xls *Excel) GetData(sheetName string, start, end string) ([]map[string]any, error) {
	rows, err := xls.Get(sheetName, start, end)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil
	}

	fields := make([]string, 0)
	for i, v := range rows[0] {
		name := cast.To[string](v)
		if name == "" {
			name = MakeColID(i)
		}
		fields = append(fields, name)
	}

	data := make([]map[string]any, 0)
	for i := 1; i < len(rows); i++ {
		rowMap := map[string]any{}
		for j, field := range fields {
			if j < len(rows[i]) {
				rowMap[field] = rows[i][j]
			} else {
				rowMap[field] = nil
			}
		}
		data = append(data, rowMap)
	}
	return data, nil
}

// 辅助方法：获取或创建工作表
func (xls *Excel) getOrCreateSheet(name string) string {
	if name == "" {
		name = "Sheet1"
	}
	// 如果是数字索引
	if matched, _ := regexp.MatchString(`^\d+$`, name); matched {
		idx := cast.To[int](name)
		return xls.excel.GetSheetName(idx)
	}
	
	idx, _ := xls.excel.GetSheetIndex(name)
	if idx == -1 {
		xls.excel.NewSheet(name)
	}
	return name
}

// MakeCellID 根据行列索引生成单元格 ID（如 0, 0 -> "A1"）。
func MakeCellID(col, row int) string {
	return MakeColID(col) + strconv.Itoa(row+1)
}

// MakeColID 根据列索引生成列 ID（如 0 -> "A", 26 -> "AA"）。
func MakeColID(col int) string {
	colName := ""
	for col >= 0 {
		colName = string(rune(col%26+65)) + colName
		col = col/26 - 1
	}
	return colName
}

// ParseCellID 解析单元格 ID（如 "A1" -> 0, 0）。
func ParseCellID(cell string) (col, row int) {
	if cell == "" {
		return 0, 0
	}
	
	// 找到第一个数字的位置
	numIdx := -1
	for i, r := range cell {
		if r >= '0' && r <= '9' {
			numIdx = i
			break
		}
	}

	if numIdx == -1 {
		// 只有字母，当做列处理
		return parseCol(cell), 0
	}

	col = parseCol(cell[:numIdx])
	row, _ = strconv.Atoi(cell[numIdx:])
	row-- // 转为 0 索引
	if row < 0 {
		row = 0
	}
	return col, row
}

func parseCol(colStr string) int {
	col := 0
	for _, r := range colStr {
		if r >= 'A' && r <= 'Z' {
			col = col*26 + int(r-'A'+1)
		} else if r >= 'a' && r <= 'z' {
			col = col*26 + int(r-'a'+1)
		}
	}
	return col - 1
}

var floatMatcher = regexp.MustCompile(`^[\d.]{1,18}$`)

func isFloat(v any) bool {
	s := cast.To[string](v)
	return bytes.ContainsRune([]byte(s), '.') && floatMatcher.MatchString(s)
}
