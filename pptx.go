package document

import (
	"fmt"
	"io"

	"apigo.cc/go/cast"
	"apigo.cc/go/file"
	"github.com/young2j/oxmltotext/pptxtotext"
)

// Pptx 封装了 PowerPoint 文档的读取与识别。
type Pptx struct {
	filename string
	Content  string
	Metadata map[string]any
}

// OpenPptx 打开一个 PowerPoint 文档 (.pptx)。
func OpenPptx(filename string) (*Pptx, error) {
	if !file.Exists(filename) {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	p := &Pptx{
		filename: filename,
		Metadata: make(map[string]any),
	}
	
	pp, err := pptxtotext.Open(filename)
	if err == nil {
		defer pp.Close()
		p.Content, _ = pp.ExtractTexts()
	}
	
	return p, nil
}

// ToJSON 返回结构化 JSON。
func (p *Pptx) ToJSON() string {
	res, _ := cast.ToJSON(map[string]any{
		"metadata": p.Metadata,
		"content":  p.Content,
	})
	return res
}

// ToMarkdown 返回 Markdown。
func (p *Pptx) ToMarkdown() string {
	return p.Content
}

// Save 保存文档（目前保存为提取后的文本）。
func (p *Pptx) Save(filename ...string) error {
	path := p.filename
	if len(filename) > 0 && filename[0] != "" {
		path = filename[0]
	}
	return file.Write(path, p.Content)
}

// ReadText 从 io.Reader 中读取并提取 PPT 文本。
func (p *Pptx) ReadText(r io.ReaderAt, size int64) (string, error) {
	pp, err := pptxtotext.OpenReader(r, size)
	if err != nil {
		return "", err
	}
	defer pp.Close()

	return pp.ExtractTexts()
}
