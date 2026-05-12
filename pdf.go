package office

import (
	"bytes"
	"io"
	"os"
	"strings"

	"apigo.cc/go/file"
	"github.com/dslipak/pdf"
)

// PDF 封装了 PDF 文档的读取操作。
type PDF struct {
	filename string
}

// OpenPDF 打开一个 PDF 文档。
func OpenPDF(filename string) (*PDF, error) {
	if !file.Exists(filename) {
		return nil, os.ErrNotExist
	}
	return &PDF{filename: filename}, nil
}

// Text 提取 PDF 中的所有文本。
func (p *PDF) Text() (string, error) {
	f, err := pdf.Open(p.filename)
	if err != nil {
		return "", err
	}
	
	var b bytes.Buffer
	t, err := f.GetPlainText()
	if err != nil {
		return "", err
	}
	
	_, err = io.Copy(&b, t)
	if err != nil {
		return "", err
	}
	
	return b.String(), nil
}

// ToMarkdown 将 PDF 内容转换为 Markdown 格式。
func (p *PDF) ToMarkdown() (string, error) {
	text, err := p.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}

// Info 获取 PDF 的元数据。
func (p *PDF) Info() map[string]any {
	f, err := pdf.Open(p.filename)
	if err != nil {
		return nil
	}
	
	info := make(map[string]any)
	trailer := f.Trailer()
	infoDict := trailer.Key("Info")
	if !infoDict.IsNull() {
		for _, field := range infoDict.Keys() {
			val := infoDict.Key(field).Text()
			if val != "" {
				info[strings.ToLower(field)] = val
			}
		}
	}
	info["pages"] = f.NumPage()
	return info
}
