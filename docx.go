package document

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"apigo.cc/go/cast"
	"apigo.cc/go/file"
	"github.com/young2j/oxmltotext/docxtotext"
)

// Docx 封装了 Word 文档的读取与识别。
type Docx struct {
	filename string
	Content  string
	Metadata map[string]any
}

// OpenDocx 打开一个 Word 文档 (.docx)。
func OpenDocx(filename string) (*Docx, error) {
	if !file.Exists(filename) {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	d := &Docx{
		filename: filename,
		Metadata: make(map[string]any),
	}

	// 尝试提取层级结构的 Markdown
	md, err := d.extractMarkdown(filename)
	if err == nil && md != "" {
		d.Content = md
	} else {
		// 回退到纯文本提取
		dp, err := docxtotext.Open(filename)
		if err == nil {
			defer dp.Close()
			d.Content, _ = dp.ExtractTexts()
		}
	}

	return d, nil
}

// extractMarkdown 尝试从 docx 的 XML 中提取带标题和表格的 Markdown。
func (d *Docx) extractMarkdown(filename string) (string, error) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var docXML io.ReadCloser
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			docXML, err = f.Open()
			break
		}
	}
	if docXML == nil {
		return "", fmt.Errorf("word/document.xml not found")
	}
	defer docXML.Close()

	decoder := xml.NewDecoder(docXML)
	var sb strings.Builder
	var inT bool
	var currentStyle string
	var inTable bool
	var tableRows [][]string
	var currentRow []string
	var cellText strings.Builder

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch se := t.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "p":
				currentStyle = ""
				cellText.Reset()
			case "pStyle":
				for _, attr := range se.Attr {
					if attr.Name.Local == "val" {
						currentStyle = attr.Value
					}
				}
			case "t":
				inT = true
			case "tbl":
				inTable = true
				tableRows = nil
			case "tr":
				currentRow = nil
			case "tc":
				cellText.Reset()
			}
		case xml.EndElement:
			switch se.Name.Local {
			case "p":
				if inTable {
					// Paragraph inside table cell is handled by cellEnd
				} else {
					sb.WriteString("\n")
				}
			case "t":
				inT = false
			case "tc":
				currentRow = append(currentRow, strings.TrimSpace(cellText.String()))
				cellText.Reset()
			case "tr":
				tableRows = append(tableRows, currentRow)
			case "tbl":
				inTable = false
				sb.WriteString(renderMarkdownTable(tableRows))
				sb.WriteString("\n")
			}
		case xml.CharData:
			text := string(se)
			if inTable {
				cellText.WriteString(text)
			} else if inT {
				if strings.Contains(strings.ToLower(currentStyle), "heading") || 
				   strings.Contains(strings.ToLower(currentStyle), "title") ||
				   strings.Contains(strings.ToLower(currentStyle), "subject") {
					level := "1"
					if strings.Contains(strings.ToLower(currentStyle), "heading") && len(currentStyle) > 7 {
						level = currentStyle[7:]
					}
					l := cast.To[int](level)
					if l == 0 {
						l = 1
					}
					prefix := strings.Repeat("#", l)
					sb.WriteString("\n" + prefix + " " + text + "\n")
				} else {
					sb.WriteString(text)
				}
			}
		}
	}

	return strings.TrimSpace(sb.String()), nil
}

func renderMarkdownTable(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n")
	for i, row := range rows {
		sb.WriteString("| ")
		for _, col := range row {
			sb.WriteString(strings.ReplaceAll(col, "|", "\\|"))
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

// ToJSON 返回包含元数据和内容的 JSON 字符串。
func (d *Docx) ToJSON() string {
	res, _ := cast.ToJSON(map[string]any{
		"metadata": d.Metadata,
		"content":  d.Content,
	})
	return res
}

// ToMarkdown 返回 Markdown 格式的内容。
func (d *Docx) ToMarkdown() string {
	return d.Content
}

// Save 保存文档。目前主要支持保存提取后的文本。
func (d *Docx) Save(filename ...string) error {
	path := d.filename
	if len(filename) > 0 && filename[0] != "" {
		path = filename[0]
	}
	return file.Write(path, d.Content)
}

// ReadText 从 io.Reader 中读取并提取 Word 文本。
func (d *Docx) ReadText(r io.ReaderAt, size int64) (string, error) {
	dp, err := docxtotext.OpenReader(r, size)
	if err != nil {
		return "", err
	}
	defer dp.Close()

	return dp.ExtractTexts()
}
