package document

import (
	"fmt"
	"sort"
	"strings"

	"apigo.cc/go/cast"
	"apigo.cc/go/file"
	"github.com/dslipak/pdf"
)

// PDF 封装了 PDF 文档的读取与识别。
type PDF struct {
	filename string
	Content  string
	Metadata map[string]any
}

// OpenPDF 打开一个 PDF 文档。
func OpenPDF(filename string) (*PDF, error) {
	if !file.Exists(filename) {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	p := &PDF{
		filename: filename,
		Metadata: make(map[string]any),
	}

	f, err := pdf.Open(filename)
	if err != nil {
		return nil, err
	}

	p.Metadata["pages"] = f.NumPage()
	trailer := f.Trailer()
	infoDict := trailer.Key("Info")
	if !infoDict.IsNull() {
		for _, field := range infoDict.Keys() {
			val := infoDict.Key(field).Text()
			if val != "" {
				p.Metadata[strings.ToLower(field)] = val
			}
		}
	}

	// 收集所有页面的 block
	type Block struct {
		Type     string
		Level    int
		Text     string
		FontSize float64
	}
	var allBlocks []Block

	for i := 1; i <= f.NumPage(); i++ {
		page := f.Page(i)
		if page.V.IsNull() {
			continue
		}

		content := page.Content()
		texts := content.Text
		if len(texts) == 0 {
			continue
		}

		// 1. 估算正文字体大小（众数）
		fontSizes := make(map[int]int)
		for _, t := range texts {
			fontSizes[int(t.FontSize)]++
		}
		bodySize := 0
		maxCount := 0
		for size, count := range fontSizes {
			if count > maxCount {
				maxCount = count
				bodySize = size
			}
		}

		// 2. 按行分组（基于 Y 坐标）
		type Line struct {
			Y           float64
			MaxFontSize float64
			Text        string
			Texts       []pdf.Text
		}
		var lines []Line
		for _, t := range texts {
			found := false
			for i := range lines {
				if t.Y > lines[i].Y-2 && t.Y < lines[i].Y+2 {
					lines[i].Texts = append(lines[i].Texts, t)
					found = true
					break
				}
			}
			if !found {
				lines = append(lines, Line{Y: t.Y, Texts: []pdf.Text{t}})
			}
		}

		sort.Slice(lines, func(i, j int) bool {
			return lines[i].Y > lines[j].Y
		})

		for i := range lines {
			sort.Slice(lines[i].Texts, func(m, n int) bool {
				return lines[i].Texts[m].X < lines[i].Texts[n].X
			})
			var sb strings.Builder
			maxFS := 0.0
			for _, t := range lines[i].Texts {
				sb.WriteString(t.S)
				if t.FontSize > maxFS {
					maxFS = t.FontSize
				}
			}
			lines[i].Text = strings.TrimSpace(sb.String())
			lines[i].MaxFontSize = maxFS
		}

		// 3. 语义块识别
		for j := 0; j < len(lines); j++ {
			line := lines[j]
			if line.Text == "" {
				continue
			}

			isTableLine, cells := p.identifyTableLine(line)
			if isTableLine {
				tableStr := "| " + strings.Join(cells, " | ") + " |"
				allBlocks = append(allBlocks, Block{Type: "table", Text: tableStr})
				continue
			}

			if line.MaxFontSize > float64(bodySize)+1 {
				level := 1
				if line.MaxFontSize < float64(bodySize)+4 {
					level = 3
				} else if line.MaxFontSize < float64(bodySize)+8 {
					level = 2
				}

				fullText := line.Text
				lastY := line.Y
				for k := j + 1; k < len(lines); k++ {
					if lines[k].Text != "" && MathAbs(lines[k].MaxFontSize-line.MaxFontSize) < 1.0 {
						if MathAbs(lines[k].Y-lastY) < 25 {
							fullText += " " + lines[k].Text
							lastY = lines[k].Y
							j = k
						} else {
							break
						}
					} else {
						break
					}
				}

				if j < 5 && i == 1 {
					if title, ok := p.Metadata["title"].(string); ok && strings.Contains(title, fullText) {
						continue
					}
				}

				if len(allBlocks) > 0 && allBlocks[len(allBlocks)-1].Type == "heading" && allBlocks[len(allBlocks)-1].Level == level {
					allBlocks[len(allBlocks)-1].Text += " " + fullText
				} else {
					allBlocks = append(allBlocks, Block{Type: "heading", Level: level, Text: fullText})
				}
			} else {
				if (line.Y < 50 || line.Y > 800) && isPageNumber(line.Text) {
					continue
				}

				fullText := line.Text
				lastY := line.Y
				for k := j + 1; k < len(lines); k++ {
					if lines[k].Text != "" && lines[k].MaxFontSize <= float64(bodySize)+1 {
						isT, _ := p.identifyTableLine(lines[k])
						if isT {
							break
						}
						if (lines[k].Y < 50 || lines[k].Y > 800) && isPageNumber(lines[k].Text) {
							break
						}
						if MathAbs(lines[k].Y-lastY) > 25 {
							break
						}

						fullText += lines[k].Text
						lastY = lines[k].Y
						j = k

						if isPunctuation(lines[k].Text) {
							break
						}
					} else {
						break
					}
				}
				allBlocks = append(allBlocks, Block{Type: "paragraph", Text: fullText})
			}
		}
	}

	// 4. 智能封面/标题逻辑 (全局)
	h1Count := 0
	for _, b := range allBlocks {
		if b.Type == "heading" && b.Level == 1 {
			h1Count++
		}
	}
	shouldPromote := h1Count <= 1

	contentStarted := false
	hasMetadataTitle := false
	if t, ok := p.Metadata["title"].(string); ok && t != "" {
		hasMetadataTitle = true
	}

	firstHeadingProcessed := false

	var sb strings.Builder
	for _, b := range allBlocks {
		if b.Type == "heading" {
			level := b.Level
			
			if !contentStarted {
				if !firstHeadingProcessed {
					firstHeadingProcessed = true
					// 如果有 Metadata Title，则 Metadata Title 充当了真正的第一级标题
					// 我们把遇到的第一个大标题降级为正文文本（除非它已经是标准章节）
					if hasMetadataTitle && !isStandardSection(b.Text) {
						sb.WriteString("\n" + b.Text + "\n\n")
						continue
					} else {
						// 否则作为文档的主标题
						sb.WriteString("\n# " + b.Text + "\n\n")
						continue
					}
				} else {
					if isStandardSection(b.Text) {
						contentStarted = true
						sb.WriteString("\n# " + b.Text + "\n\n")
						continue
					}
					// 否则作为封面副标题/文本
					sb.WriteString(b.Text + "\n\n")
					continue
				}
			}

			if shouldPromote {
				if level == 2 {
					level = 1
				} else if level == 3 {
					level = 2
				}
			}
			sb.WriteString("\n" + strings.Repeat("#", level) + " " + b.Text + "\n\n")
		} else if b.Type == "paragraph" {
			contentStarted = true
			sb.WriteString(b.Text + "\n\n")
		} else if b.Type == "table" {
			contentStarted = true
			sb.WriteString(b.Text + "\n")
		}
	}

	p.Content = strings.TrimSpace(sb.String())
	return p, nil
}

func isStandardSection(s string) bool {
	s = strings.TrimSpace(s)
	// 常见的章节开头关键词
	standards := []string{"引言", "摘要", "目录", "前言", "结论", "背景", "概述", "Introduction", "Abstract", "Conclusion", "Summary"}
	for _, std := range standards {
		if strings.Contains(s, std) {
			return true
		}
	}
	// 如果标题带有数字编号且较短，也认为是章节开始
	if len([]rune(s)) < 20 {
		if s[0] >= '0' && s[0] <= '9' {
			return true
		}
		if strings.HasPrefix(s, "第") && (strings.Contains(s, "章") || strings.Contains(s, "节") || strings.Contains(s, "部分")) {
			return true
		}
	}
	return false
}

func isPageNumber(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	isNum := true
	for _, r := range s {
		if r < '0' || r > '9' {
			isNum = false
			break
		}
	}
	if isNum {
		return true
	}
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "page") || (strings.HasPrefix(s, "-") && strings.HasSuffix(s, "-"))
}

func isPunctuation(s string) bool {
	if s == "" {
		return false
	}
	runes := []rune(s)
	last := runes[len(runes)-1]
	return strings.ContainsRune("。.!！?？:：", last)
}

func (p *PDF) identifyTableLine(line struct {
	Y           float64
	MaxFontSize float64
	Text        string
	Texts       []pdf.Text
}) (bool, []string) {
	if len(line.Texts) < 3 {
		return false, nil
	}
	var cells []string
	var currentCell strings.Builder
	lastX := -1.0
	for _, t := range line.Texts {
		if lastX != -1.0 && t.X-(lastX) > 40 {
			content := strings.TrimSpace(currentCell.String())
			if content != "" {
				cells = append(cells, content)
			}
			currentCell.Reset()
		}
		currentCell.WriteString(t.S)
		lastX = t.X + t.W
	}
	finalCell := strings.TrimSpace(currentCell.String())
	if finalCell != "" {
		cells = append(cells, finalCell)
	}

	if len(cells) >= 2 {
		allSingleChar := true
		for _, c := range cells {
			r := []rune(c)
			if len(r) > 1 {
				allSingleChar = false
			}
			if len(r) > 40 {
				return false, nil
			}
		}
		if allSingleChar {
			return false, nil
		}
		return true, cells
	}
	return false, nil
}

func MathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// ToJSON 返回结构化 JSON。
func (p *PDF) ToJSON() string {
	res, _ := cast.ToJSON(map[string]any{
		"metadata": p.Metadata,
		"content":  p.Content,
	})
	return res
}

// ToMarkdown 返回 Markdown。
func (p *PDF) ToMarkdown() string {
	if p.Content == "" {
		return ""
	}
	var sb strings.Builder
	if title, ok := p.Metadata["title"]; ok && title != "" {
		sb.WriteString("# " + cast.To[string](title) + "\n\n")
	}
	sb.WriteString(p.Content)
	return sb.String()
}

// Save 保存（目前保存为提取后的文本）。
func (p *PDF) Save(filename ...string) error {
	path := p.filename
	if len(filename) > 0 && filename[0] != "" {
		path = filename[0]
	}
	return file.Write(path, p.Content)
}
