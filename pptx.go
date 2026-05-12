package office

import (
	"bytes"
	"io"

	"apigo.cc/go/file"
	"github.com/young2j/oxmltotext/pptxtotext"
)

// Pptx 封装了 PowerPoint 文档的读取操作。
type Pptx struct {
	filename string
}

// OpenPptx 打开一个 PowerPoint 文档 (.pptx)。
func OpenPptx(filename string) (*Pptx, error) {
	if !file.Exists(filename) {
		return nil, file.ErrNotExist
	}
	return &Pptx{filename: filename}, nil
}

// Text 提取文档中的所有文本。
func (p *Pptx) Text() (string, error) {
	f, err := file.Open(p.filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return p.ReadText(f)
}

// ReadText 从 io.Reader 中读取并提取 PPT 文本。
func (p *Pptx) ReadText(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	
	res, err := pptxtotext.Extract(bytes.NewReader(data), nil)
	if err != nil {
		return "", err
	}
	return res, nil
}
