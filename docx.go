package office

import (
	"bytes"
	"io"

	"apigo.cc/go/file"
	"github.com/young2j/oxmltotext/docxtotext"
)

// Docx 封装了 Word 文档的读取操作。
type Docx struct {
	filename string
}

// OpenDocx 打开一个 Word 文档 (.docx)。
func OpenDocx(filename string) (*Docx, error) {
	if !file.Exists(filename) {
		return nil, file.ErrNotExist
	}
	return &Docx{filename: filename}, nil
}

// Text 提取文档中的所有文本。
func (d *Docx) Text() (string, error) {
	f, err := file.Open(d.filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return d.ReadText(f)
}

// ReadText 从 io.Reader 中读取并提取 Word 文本。
func (d *Docx) ReadText(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	
	res, err := docxtotext.Extract(bytes.NewReader(data), nil)
	if err != nil {
		return "", err
	}
	return res, nil
}
