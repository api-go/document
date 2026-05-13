package document

import (
	"strings"
	"testing"

	"apigo.cc/go/file"
)

func TestGraph_Basic(t *testing.T) {
	filename := "test.graph"
	defer file.Remove(filename)

	g := NewGraph()
	g.Title = "Novel Plot"

	g.AddNode(&Node{
		ID:      "1",
		Title:   "Chapter 1",
		Content: "Alice meets Bob.",
		Type:    "scene",
		Links:   []string{"2"},
	})

	g.AddNode(&Node{
		ID:      "2",
		Title:   "Chapter 2",
		Content: "They go to the park.",
		Type:    "scene",
		Parents: []string{"1"},
	})

	err := g.Save(filename)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	g2, err := OpenGraph(filename)
	if err != nil {
		t.Fatalf("OpenGraph failed: %v", err)
	}

	if g2.Title != "Novel Plot" {
		t.Errorf("Expected Title 'Novel Plot', got '%s'", g2.Title)
	}

	if len(g2.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(g2.Nodes))
	}

	md := g2.ToMarkdown()
	if !strings.Contains(md, "mermaid") {
		t.Errorf("Markdown should contain mermaid diagram")
	}
	if !strings.Contains(md, "Chapter 1") {
		t.Errorf("Markdown should contain Chapter 1")
	}

	jsonStr := g2.ToJSON()
	if !strings.Contains(jsonStr, "Novel Plot") {
		t.Errorf("JSON should contain Title")
	}
}

func TestGraph_Unified(t *testing.T) {
	filename := "unified.graph"
	defer file.Remove(filename)

	doc, _ := Open(filename)
	if g, ok := doc.(*Graph); ok {
		g.Title = "Unified Test"
		g.AddNode(&Node{Title: "Root"})
		doc.Save()
	}

	doc2, _ := Open(filename)
	if doc2.ToJSON() == "" {
		t.Errorf("ToJSON failed")
	}
}
