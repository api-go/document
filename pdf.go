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

	var sb strings.Builder
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

		// 处理页面内容
		sb.WriteString(p.processPageTexts(texts))
	}
	p.Content = strings.TrimSpace(sb.String())

	return p, nil
}

func (p *PDF) processPageTexts(texts []pdf.Text) string {
	if len(texts) == 0 {
		return ""
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

	// 按 Y 降序排列（从上到下）
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Y > lines[j].Y
	})

	// 预处理每一行的文本和最大字体
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
	type Block struct {
		Type     string // heading, paragraph, table
		Level    int    // for heading
		Text     string
		FontSize float64
	}
	var blocks []Block
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line.Text == "" {
			continue
		}

		// 表格识别逻辑
		isTableLine, cells := p.identifyTableLine(line)
		if isTableLine {
			tableStr := "| " + strings.Join(cells, " | ") + " |"
			blocks = append(blocks, Block{Type: "table", Text: tableStr})
			continue
		}

		// 标题识别逻辑 (比正文大)
		if line.MaxFontSize > float64(bodySize)+1 {
			level := 1
			if line.MaxFontSize < float64(bodySize)+4 {
				level = 3
			} else if line.MaxFontSize < float64(bodySize)+8 {
				level = 2
			}

			// 合并紧随其后的同字体行（处理跨行标题）
			fullText := line.Text
			lastY := line.Y
			for j := i + 1; j < len(lines); j++ {
				if lines[j].Text != "" && MathAbs(lines[j].MaxFontSize-line.MaxFontSize) < 1.0 {
					if MathAbs(lines[j].Y-lastY) < 25 {
						fullText += " " + lines[j].Text
						lastY = lines[j].Y
						i = j
					} else {
						break
					}
				} else {
					break
				}
			}

			// 检查是否重复
			if i < 5 {
				if title, ok := p.Metadata["title"].(string); ok && strings.Contains(title, fullText) {
					continue
				}
			}

			// 合并同级标题
			if len(blocks) > 0 && blocks[len(blocks)-1].Type == "heading" && blocks[len(blocks)-1].Level == level {
				blocks[len(blocks)-1].Text += " " + fullText
			} else {
				blocks = append(blocks, Block{Type: "heading", Level: level, Text: fullText})
			}
		} else {
			// 跳过页码
			if (line.Y < 50 || line.Y > 800) && isPageNumber(line.Text) {
				continue
			}

			// 段落识别
			fullText := line.Text
			lastY := line.Y
			for j := i + 1; j < len(lines); j++ {
				if lines[j].Text != "" && lines[j].MaxFontSize <= float64(bodySize)+1 {
					isT, _ := p.identifyTableLine(lines[j])
					if isT {
						break
					}
					if (lines[j].Y < 50 || lines[j].Y > 800) && isPageNumber(lines[j].Text) {
						break
					}
					if MathAbs(lines[j].Y-lastY) > 25 {
						break
					}

					fullText += lines[j].Text
					lastY = lines[j].Y
					i = j

					if isPunctuation(lines[j].Text) {
						break
					}
				} else {
					break
				}
			}
			blocks = append(blocks, Block{Type: "paragraph", Text: fullText})
		}
	}

	// 4. 智能封面/标题逻辑
	// 策略：
	// 1. 第一个 Heading 始终是文档标题 (#)
	// 2. 在遇到第一个明确的“章节标题”或“正文段落”之前，中间的 Heading 如果很长，则转为正文文本。
	// 3. 统计 H1，如果只有一个 H1，则尝试将 H2 提升为 H1。
	
	h1Count := 0
	for _, b := range blocks {
		if b.Type == "heading" && b.Level == 1 {
			h1Count++
		}
	}
	shouldPromote := h1Count <= 1

	firstHeadingFound := false
	contentStarted := false

	var res strings.Builder
	for _, b := range blocks {
		if b.Type == "heading" {
			level := b.Level
			
			if !contentStarted {
				if !firstHeadingFound {
					// 文档总标题
					res.WriteString("\n# " + b.Text + "\n\n")
					firstHeadingFound = true
					continue
				} else {
					// 封面期间的其他标题
					// 如果是已知的章节名，则认为内容开始了
					if isStandardSection(b.Text) {
						contentStarted = true
						// 章节名也应该是 #
						res.WriteString("\n# " + b.Text + "\n\n")
						continue
					}
					// 否则作为封面副标题/文本
					res.WriteString(b.Text + "\n\n")
					continue
				}
			}

			// 内容已经开始
			if shouldPromote {
				if level == 2 {
					level = 1
				} else if level == 3 {
					level = 2
				}
			}
			res.WriteString("\n" + strings.Repeat("#", level) + " " + b.Text + "\n\n")
		} else if b.Type == "paragraph" {
			contentStarted = true // 遇到第一个正文段落，标志着封面/标题期结束
			res.WriteString(b.Text + "\n\n")
		} else if b.Type == "table" {
			contentStarted = true
			res.WriteString(b.Text + "\n")
		}
	}

	return res.String()
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
