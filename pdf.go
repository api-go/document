package office

import (
	"bytes"
	"io"
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
		return nil, file.ErrNotExist
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

// Info 获取 PDF 的元数据。
func (p *PDF) Info() map[string]any {
	f, err := pdf.Open(p.filename)
	if err != nil {
		return nil
	}
	
	info := make(map[string]any)
	// 常见的 PDF 元数据字段
	fields := []string{"Title", "Author", "Subject", "Keywords", "Creator", "Producer", "CreationDate", "ModDate"}
	for _, field := range fields {
		val := f.GetInfo().Get(field)
		if val != "" {
			info[strings.ToLower(field)] = val
		}
	}
	info["pages"] = f.NumPage()
	return info
}
