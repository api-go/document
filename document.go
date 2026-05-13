package document

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Document 定义了所有办公文档的统一行为。
type Document interface {
	// ToJSON 将文档转换为结构化 JSON 字符串。
	ToJSON() string
	// ToMarkdown 将文档内容转换为 Markdown 格式。
	ToMarkdown() string
	// Save 保存文档到文件系统。
	Save(filename ...string) error
}

// Open 根据文件后缀名自动打开文档并返回标准接口。
func Open(filename string, password ...string) (Document, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xlsx", ".xlsm":
		return OpenExcel(filename, password...)
	case ".docx":
		return OpenDocx(filename)
	case ".pptx":
		return OpenPptx(filename)
	case ".pdf":
		return OpenPDF(filename)
	case ".graph":
		return OpenGraph(filename)
	case ".csv":
		return OpenCSV(filename)
	case ".md", ".markdown", ".txt":
		return OpenMarkdown(filename)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// Create 创建指定类型的空白文档。
func Create(ext string) (Document, error) {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	switch ext {
	case "xlsx":
		return NewExcel(), nil
	case "docx":
		// TODO: 支持创建 Word
		return nil, fmt.Errorf("creating docx not supported yet")
	case "graph":
		return NewGraph(), nil
	case "csv":
		return &CSV{}, nil
	case "md", "markdown":
		return &Markdown{}, nil
	default:
		return nil, fmt.Errorf("unsupported creation type: %s", ext)
	}
}
