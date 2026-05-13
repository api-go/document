package document

import (
	"encoding/csv"
	"strings"

	"apigo.cc/go/cast"
	"apigo.cc/go/file"
)

// CSV 封装了 CSV 文件的读写。
type CSV struct {
	filename string
	Data     [][]string
}

// OpenCSV 打开一个 CSV 文件。
func OpenCSV(filename string) (*CSV, error) {
	if !file.Exists(filename) {
		return &CSV{filename: filename}, nil
	}
	
	content, err := file.Read(filename)
	if err != nil {
		return nil, err
	}
	
	r := csv.NewReader(strings.NewReader(content))
	data, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	
	return &CSV{
		filename: filename,
		Data:     data,
	}, nil
}

// ToJSON 返回 JSON 数组。
func (c *CSV) ToJSON() string {
	if len(c.Data) == 0 {
		return "[]"
	}
	
	header := c.Data[0]
	var res []map[string]string
	for i := 1; i < len(c.Data); i++ {
		row := make(map[string]string)
		for j, h := range header {
			if j < len(c.Data[i]) {
				row[h] = c.Data[i][j]
			}
		}
		res = append(res, row)
	}
	
	jsonStr, _ := cast.ToJSON(res)
	return jsonStr
}

// ToMarkdown 返回 Markdown 表格。
func (c *CSV) ToMarkdown() string {
	if len(c.Data) == 0 {
		return ""
	}
	
	var sb strings.Builder
	for i, row := range c.Data {
		sb.WriteString("| ")
		for _, col := range row {
			val := strings.ReplaceAll(col, "|", "\\|")
			sb.WriteString(val)
			sb.WriteString(" | ")
		}
		sb.WriteString("\n")
		if i == 0 {
			sb.WriteString("|")
			for range row {
				sb.WriteString(" --- |")
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// Save 保存为 CSV 文件。
func (c *CSV) Save(filename ...string) error {
	path := c.filename
	if len(filename) > 0 && filename[0] != "" {
		path = filename[0]
	}
	
	var sb strings.Builder
	w := csv.NewWriter(&sb)
	w.WriteAll(c.Data)
	w.Flush()
	
	return file.Write(path, sb.String())
}
