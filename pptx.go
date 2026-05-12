package office

import (
	"io"
	"os"

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
		return nil, os.ErrNotExist
	}
	return &Pptx{filename: filename}, nil
}

// Text 提取文档中的所有文本。
func (p *Pptx) Text() (string, error) {
	pp, err := pptxtotext.Open(p.filename)
	if err != nil {
		return "", err
	}
	defer pp.Close()

	return pp.ExtractTexts()
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

// ToMarkdown 将 PowerPoint 文档内容转换为 Markdown 格式。
func (p *Pptx) ToMarkdown() (string, error) {
	text, err := p.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}
