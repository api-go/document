package document

import (
	"archive/zip"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"path"
	"path/filepath"
	"strings"

	"github.com/api-go/cast"
	"github.com/api-go/file"
	"github.com/young2j/oxmltotext/docxtotext"
)

// Docx 封装了 Word 文档的读取与识别。
type Docx struct {
	filename      string
	Content       string
	assetMarkdown string
	assets        []Asset
	Metadata      map[string]any
}

// OpenDocx 打开一个 Word 文档 (.docx)。
func OpenDocx(filename string) (*Docx, error) {
	if !file.Exists(filename) {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	d := &Docx{
		filename: filename,
		Metadata: make(map[string]any),
	}

	// 尝试提取层级结构的 Markdown
	md, assets, err := d.extractMarkdown(filename)
	if err == nil && md != "" {
		d.assetMarkdown, d.assets = md, assets
		d.Content = inlineDocxAssets(md, assets)
	} else {
		// 回退到纯文本提取
		dp, err := docxtotext.Open(filename)
		if err == nil {
			defer dp.Close()
			d.Content, _ = dp.ExtractTexts()
		}
	}

	return d, nil
}

// extractMarkdown 尝试从 docx 的 XML 中提取带标题和表格的 Markdown。
func (d *Docx) extractMarkdown(filename string) (string, []Asset, error) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return "", nil, err
	}
	defer r.Close()

	var docXML io.ReadCloser
	files := make(map[string]*zip.File, len(r.File))
	for _, f := range r.File {
		files[f.Name] = f
		if f.Name == "word/document.xml" {
			docXML, err = f.Open()
		}
	}
	if docXML == nil {
		return "", nil, fmt.Errorf("word/document.xml not found")
	}
	defer docXML.Close()
	images := docxRelationshipImages(files)

	decoder := xml.NewDecoder(docXML)
	var sb strings.Builder
	var inT bool
	var currentStyle string
	var inTable bool
	var tableRows [][]string
	var currentRow []string
	var cellText strings.Builder
	var assets []Asset
	seenAssets := map[string]bool{}

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil, err
		}

		switch se := t.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "p":
				currentStyle = ""
				cellText.Reset()
			case "pStyle":
				for _, attr := range se.Attr {
					if attr.Name.Local == "val" {
						currentStyle = attr.Value
					}
				}
			case "t":
				inT = true
			case "tbl":
				inTable = true
				tableRows = nil
			case "tr":
				currentRow = nil
			case "tc":
				cellText.Reset()
			case "blip", "imagedata":
				for _, attr := range se.Attr {
					if attr.Name.Local != "embed" && attr.Name.Local != "id" {
						continue
					}
					if image, ok := images[attr.Value]; ok {
						if !seenAssets[image.Asset.Name] {
							assets = append(assets, image.Asset)
							seenAssets[image.Asset.Name] = true
						}
						if inTable {
							cellText.WriteString(image.Reference)
						} else {
							sb.WriteString("\n\n" + image.Reference + "\n\n")
						}
					}
					break
				}
			}
		case xml.EndElement:
			switch se.Name.Local {
			case "p":
				if inTable {
					// Paragraph inside table cell is handled by cellEnd
				} else {
					sb.WriteString("\n")
				}
			case "t":
				inT = false
			case "tc":
				currentRow = append(currentRow, strings.TrimSpace(cellText.String()))
				cellText.Reset()
			case "tr":
				tableRows = append(tableRows, currentRow)
			case "tbl":
				inTable = false
				sb.WriteString(renderMarkdownTable(tableRows))
				sb.WriteString("\n")
			}
		case xml.CharData:
			text := string(se)
			if inTable {
				cellText.WriteString(text)
			} else if inT {
				if strings.Contains(strings.ToLower(currentStyle), "heading") ||
					strings.Contains(strings.ToLower(currentStyle), "title") ||
					strings.Contains(strings.ToLower(currentStyle), "subject") {
					level := "1"
					if strings.Contains(strings.ToLower(currentStyle), "heading") && len(currentStyle) > 7 {
						level = currentStyle[7:]
					}
					l := cast.To[int](level)
					if l == 0 {
						l = 1
					}
					prefix := strings.Repeat("#", l)
					sb.WriteString("\n" + prefix + " " + text + "\n")
				} else {
					sb.WriteString(text)
				}
			}
		}
	}

	return strings.TrimSpace(sb.String()), assets, nil
}

type docxImage struct {
	Reference string
	Asset     Asset
}

func docxRelationshipImages(files map[string]*zip.File) map[string]docxImage {
	rels := files["word/_rels/document.xml.rels"]
	if rels == nil {
		return nil
	}
	r, err := rels.Open()
	if err != nil {
		return nil
	}
	defer r.Close()
	decoder := xml.NewDecoder(r)
	images := map[string]docxImage{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return images
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "Relationship" {
			continue
		}
		var id, target, relationshipType string
		for _, attr := range start.Attr {
			switch attr.Name.Local {
			case "Id":
				id = attr.Value
			case "Target":
				target = attr.Value
			case "Type":
				relationshipType = attr.Value
			}
		}
		if id == "" || target == "" || !strings.HasSuffix(strings.ToLower(relationshipType), "/image") {
			continue
		}
		name := path.Clean(path.Join("word", target))
		imageFile := files[name]
		if imageFile == nil {
			continue
		}
		imageReader, err := imageFile.Open()
		if err != nil {
			continue
		}
		data, readErr := io.ReadAll(imageReader)
		_ = imageReader.Close()
		if readErr != nil || len(data) == 0 {
			continue
		}
		contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(name)))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		assetName := path.Base(name)
		images[id] = docxImage{
			Reference: fmt.Sprintf("![%s](doc-asset://%s)", assetName, assetName),
			Asset:     Asset{Name: assetName, ContentType: contentType, Alt: assetName, Data: data},
		}
	}
	return images
}

func inlineDocxAssets(markdown string, assets []Asset) string {
	for _, asset := range assets {
		dataURL := fmt.Sprintf("data:%s;base64,%s", asset.ContentType, base64.StdEncoding.EncodeToString(asset.Data))
		markdown = strings.ReplaceAll(markdown, "doc-asset://"+asset.Name, dataURL)
	}
	return markdown
}

func renderMarkdownTable(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n")
	for i, row := range rows {
		sb.WriteString("| ")
		for _, col := range row {
			sb.WriteString(strings.ReplaceAll(col, "|", "\\|"))
			sb.WriteString(" | ")
		}
		sb.WriteString("\n")
		if i == 0 {
			sb.WriteString("|")
			for range row {
				sb.WriteString(" --- |")
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
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

// ToMarkdownWithAssets returns lightweight Markdown plus the binary resources
// it references. Callers can persist the assets and replace doc-asset URLs.
func (d *Docx) ToMarkdownWithAssets() (string, []Asset) {
	markdown := d.assetMarkdown
	if markdown == "" {
		markdown = d.Content
	}
	assets := make([]Asset, len(d.assets))
	copy(assets, d.assets)
	return markdown, assets
}

// Save 保存文档。目前主要支持保存提取后的文本。
func (d *Docx) Save(filename ...string) error {
	path := d.filename
	if len(filename) > 0 && filename[0] != "" {
		path = filename[0]
	}
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
