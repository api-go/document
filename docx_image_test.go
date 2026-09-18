package document

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocxMarkdownIncludesEmbeddedImage(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "image.docx")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(file)
	entries := map[string]string{
		"word/document.xml":            `<w:document xmlns:w="w" xmlns:a="a" xmlns:r="r"><w:body><w:p><w:r><w:t>绘制代码如下图:</w:t></w:r><w:r><a:blip r:embed="rId5"/></w:r></w:p></w:body></w:document>`,
		"word/_rels/document.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId5" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/code.png"/></Relationships>`,
		"word/media/code.png":          "PNG-CONTENT",
	}
	for name, content := range entries {
		writer, createErr := zw.Create(name)
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, writeErr := writer.Write([]byte(content)); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}

	doc, err := OpenDocx(filename)
	if err != nil {
		t.Fatal(err)
	}
	markdown := doc.ToMarkdown()
	if !strings.Contains(markdown, "绘制代码如下图:") || !strings.Contains(markdown, "![code.png](data:image/png;base64,") {
		t.Fatalf("embedded image missing from Markdown: %s", markdown)
	}
	assetMarkdown, assets := doc.ToMarkdownWithAssets()
	if !strings.Contains(assetMarkdown, "![code.png](doc-asset://code.png)") || len(assets) != 1 || string(assets[0].Data) != "PNG-CONTENT" {
		t.Fatalf("asset Markdown mismatch: markdown=%s assets=%+v", assetMarkdown, assets)
	}
}
