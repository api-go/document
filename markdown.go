package document

import (
	"fmt"

	"github.com/api-go/cast"
	"github.com/api-go/file"
)

// Markdown 封装了纯 Markdown 文档的读写，使其符合 Document 接口。
type Markdown struct {
	filename string
	Content  string
}

// OpenMarkdown 打开一个 Markdown 文件。
func OpenMarkdown(filename string) (*Markdown, error) {
	m := &Markdown{filename: filename}
	if file.Exists(filename) {
		content, err := file.Read(filename)
		if err != nil {
			return nil, err
		}
		m.Content = content
	}
	return m, nil
}

// ToJSON 返回包含内容的 JSON 字符串。
func (m *Markdown) ToJSON() string {
	res, _ := cast.ToJSON(map[string]string{
		"content": m.Content,
	})
	return res
}

// ToMarkdown 直接返回其内容。
func (m *Markdown) ToMarkdown() string {
	return m.Content
}

// Save 保存为 Markdown 文件。
func (m *Markdown) Save(filename ...string) error {
	path := m.filename
	if len(filename) > 0 && filename[0] != "" {
		path = filename[0]
	}
	if path == "" {
		return fmt.Errorf("no filename specified")
	}
	return file.Write(path, m.Content)
}
