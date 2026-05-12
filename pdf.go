package document

import (
	"bytes"
	"io"
	"os"
	"strings"

	"apigo.cc/go/cast"
	"apigo.cc/go/file"
	"github.com/dslipak/pdf"
)

// PDF 封装了 PDF 文档的读取与识别。
type PDF struct {
	filename string
	Content  string
	Metadata map[string]any
}

// OpenPDF 打开一个 PDF 文档。
func OpenPDF(filename string) (*PDF, error) {
	if !file.Exists(filename) {
		return nil, os.ErrNotExist
	}
	p := &PDF{
		filename: filename,
		Metadata: make(map[string]any),
	}
	
	f, err := pdf.Open(filename)
	if err == nil {
		p.Metadata["pages"] = f.NumPage()
		trailer := f.Trailer()
		infoDict := trailer.Key("Info")
		if !infoDict.IsNull() {
			for _, field := range infoDict.Keys() {
				val := infoDict.Key(field).Text()
				if val != "" {
					p.Metadata[strings.ToLower(field)] = val
				}
			}
		}
		
		var b bytes.Buffer
		if t, err := f.GetPlainText(); err == nil {
			io.Copy(&b, t)
			p.Content = b.String()
		}
	}
	
	return p, nil
}

// ToJSON 返回结构化 JSON。
func (p *PDF) ToJSON() string {
	res, _ := cast.ToJSON(map[string]any{
		"metadata": p.Metadata,
		"content":  p.Content,
	})
	return res
}

// ToMarkdown 返回 Markdown。
func (p *PDF) ToMarkdown() string {
	return p.Content
}

// Save 保存（目前保存为提取后的文本）。
func (p *PDF) Save(filename ...string) error {
	path := p.filename
	if len(filename) > 0 && filename[0] != "" {
		path = filename[0]
	}
	return file.Write(path, p.Content)
}
