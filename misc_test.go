package document

import (
	"strings"
	"testing"

	"apigo.cc/go/file"
)

func TestCSV(t *testing.T) {
	filename := "test.csv"
	defer file.Remove(filename)

	c, _ := OpenCSV(filename)
	c.Data = [][]string{
		{"Name", "Age"},
		{"Alice", "25"},
		{"Bob", "30"},
	}
	c.Save()

	doc, err := Open(filename)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	md := doc.ToMarkdown()
	if !strings.Contains(md, "Alice") || !strings.Contains(md, "| --- |") {
		t.Errorf("Unexpected Markdown: %s", md)
	}

	jsonStr := doc.ToJSON()
	if !strings.Contains(jsonStr, "\"Name\":\"Alice\"") {
		t.Errorf("Unexpected JSON: %s", jsonStr)
	}
}

func TestMarkdown(t *testing.T) {
	filename := "sample_markdown.md"
	defer file.Remove(filename)

	m, _ := OpenMarkdown(filename)
	m.Content = "# Hello\nWorld"
	m.Save()

	doc, err := Open(filename)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if doc.ToMarkdown() != "# Hello\nWorld" {
		t.Errorf("Unexpected Markdown: %s", doc.ToMarkdown())
	}
}
