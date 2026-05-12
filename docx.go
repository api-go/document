package office

import (
	"io"
	"os"

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
		return nil, os.ErrNotExist
	}
	return &Docx{filename: filename}, nil
}

// Text 提取文档中的所有文本。
func (d *Docx) Text() (string, error) {
	dp, err := docxtotext.Open(d.filename)
	if err != nil {
		return "", err
	}
	defer dp.Close()

	return dp.ExtractTexts()
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

// ToMarkdown 将 Word 文档内容转换为 Markdown 格式。
func (d *Docx) ToMarkdown() (string, error) {
	text, err := d.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}
