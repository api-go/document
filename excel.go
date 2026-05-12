package office

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"apigo.cc/go/cast"
	"apigo.cc/go/file"
	"github.com/xuri/excelize/v2"
)

// Excel 封装了 Excel 文件的核心操作。
type Excel struct {
	filename string
	password string
	excel    *excelize.File
}

// NewExcel 创建一个新的 Excel 对象。
func NewExcel() *Excel {
	return &Excel{
		excel: excelize.NewFile(),
	}
}

// OpenExcel 打开一个现有的 Excel 文件。
func OpenExcel(filename string, password ...string) (*Excel, error) {
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
func (xls *Excel) Save(filename ...string) error {
	if len(filename) > 0 && filename[0] != "" {
		xls.filename = filename[0]
	}
	if xls.filename == "" {
		return fmt.Errorf("no filename specified")
	}
	return xls.excel.SaveAs(xls.filename, excelize.Options{Password: xls.password})
}

// ToJSON 将第一个工作表的数据转换为 JSON 字符串。
func (xls *Excel) ToJSON() string {
	sheets := xls.Sheets()
	if len(sheets) == 0 {
		return "[]"
	}
	data, _ := xls.GetData(sheets[0], "A1", "")
	res, _ := cast.ToJSON(data)
	return res
}

// ToMarkdown 将第一个工作表的数据转换为 Markdown 表格。
func (xls *Excel) ToMarkdown() string {
	sheets := xls.Sheets()
	if len(sheets) == 0 {
		return ""
	}
	rows, _ := xls.Get(sheets[0], "A1", "")
	if len(rows) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, row := range rows {
		sb.WriteString("| ")
		for _, col := range row {
			sb.WriteString(cast.To[string](col))
			sb.WriteString(" | ")
		}
		sb.WriteString("\n")
		if i == 0 {
			// 分隔线
			sb.WriteString("|")
			for range row {
				sb.WriteString(" --- |")
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// --- 基础操作方法 ---

// Set 设置指定单元格范围的值。
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
			default:
				rowData = append(rowData, v)
			}
		}
		result = append(result, rowData)
	}
	return result, nil
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

// Sheets 返回所有工作表名称。
func (xls *Excel) Sheets() []string {
	return xls.excel.GetSheetList()
}

// 辅助方法：获取或创建工作表
func (xls *Excel) getOrCreateSheet(name string) string {
	if name == "" {
		name = "Sheet1"
	}
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

// MakeCellID 生成单元格 ID。
func MakeCellID(col, row int) string {
	return MakeColID(col) + strconv.Itoa(row+1)
}

// MakeColID 生成列 ID。
func MakeColID(col int) string {
	colName := ""
	for col >= 0 {
		colName = string(rune(col%26+65)) + colName
		col = col/26 - 1
	}
	return colName
}

// ParseCellID 解析单元格 ID。
func ParseCellID(cell string) (col, row int) {
	if cell == "" {
		return 0, 0
	}
	numIdx := -1
	for i, r := range cell {
		if r >= '0' && r <= '9' {
			numIdx = i
			break
		}
	}
	if numIdx == -1 {
		return parseCol(cell), 0
	}
	col = parseCol(cell[:numIdx])
	row, _ = strconv.Atoi(cell[numIdx:])
	row--
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
