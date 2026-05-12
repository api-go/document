package office

import (
	"io"
	"os"

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
		return nil, os.ErrNotExist
	}
	d := &Docx{
		filename: filename,
		Metadata: make(map[string]any),
	}
	
	dp, err := docxtotext.Open(filename)
	if err == nil {
		defer dp.Close()
		d.Content, _ = dp.ExtractTexts()
	}
	
	return d, nil
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
	// 如果是原格式保存，目前仅作为文本保存或保持原样
	// 复杂的 Docx 写入暂未集成
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
