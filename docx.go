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

// extractMarkdown 尝试从 docx 的 XML 中提取带标题的 Markdown。
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
			if se.Name.Local == "p" {
				currentStyle = ""
			} else if se.Name.Local == "pStyle" {
				for _, attr := range se.Attr {
					if attr.Name.Local == "val" {
						currentStyle = attr.Value
					}
				}
			} else if se.Name.Local == "t" {
				inT = true
			}
		case xml.EndElement:
			if se.Name.Local == "p" {
				sb.WriteString("\n")
			} else if se.Name.Local == "t" {
				inT = false
			}
		case xml.CharData:
			if inT {
				text := string(se)
				if strings.Contains(strings.ToLower(currentStyle), "heading") {
					level := "1"
					if len(currentStyle) > 7 {
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
