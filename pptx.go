package document

import (
	"fmt"
	"io"
	"strings"

	"github.com/api-go/cast"
	"github.com/api-go/file"
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
	if p.Content == "" {
		return ""
	}
	
	lines := strings.Split(p.Content, "\n")
	var res []string
	nextIsMainTitle := true
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		
		// 检查幻灯片分隔符
		if strings.Contains(line, "--------------------------------") {
			res = append(res, "\n---")
			nextIsMainTitle = true
			continue
		}
		
		// 启发式识别标题：行短、无句末标点
		if len([]rune(trimmed)) < 50 && !strings.HasSuffix(trimmed, ".") && !strings.HasSuffix(trimmed, "。") && !strings.HasSuffix(trimmed, ":") && !strings.HasSuffix(trimmed, "：") {
			if nextIsMainTitle {
				res = append(res, "\n# "+trimmed)
				nextIsMainTitle = false
			} else {
				res = append(res, "\n## "+trimmed)
			}
		} else {
			res = append(res, trimmed)
			// 如果该页已经有了主标题，后续的长文本不会重置标题状态
		}
	}
	return strings.TrimSpace(strings.Join(res, "\n"))
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
