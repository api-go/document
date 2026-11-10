package document

import (
	"fmt"
	"strings"

	"github.com/api-go/cast"
	"github.com/api-go/file"
)

// Node 代表文档中的一个节点，可以是一个场景、一个角色或一个知识点。
type Node struct {
	ID      string         `json:"id"`
	Title   string         `json:"title"`
	Content string         `json:"content"`
	Type    string         `json:"type,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
	Links   []string       `json:"links,omitempty"`   // 与其他节点的关联 (ID 列表)
	Parents []string       `json:"parents,omitempty"` // 父节点 (用于层级结构)
}

// Graph 是一种具有关联关系的文档，适用于小说大纲、策划分镜、思维导图等场景。
type Graph struct {
	filename string
	Title    string           `json:"title"`
	Nodes    map[string]*Node `json:"nodes"`
}

// NewGraph 创建一个新的关系型文档。
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
	}
}

// AddNode 添加或更新一个节点。
func (g *Graph) AddNode(n *Node) {
	if n.ID == "" {
		n.ID = cast.To[string](len(g.Nodes) + 1)
	}
	g.Nodes[n.ID] = n
}

// ToJSON 返回文档的结构化 JSON 表示。
func (g *Graph) ToJSON() string {
	res, _ := cast.ToJSON(g)
	return res
}

// ToMarkdown 将关系型文档转换为带有 Mermaid 图表的 Markdown。
func (g *Graph) ToMarkdown() string {
	var sb strings.Builder
	sb.WriteString("# " + g.Title + "\n\n")

	// 生成 Mermaid 关系图
	sb.WriteString("```mermaid\ngraph TD\n")
	for id, node := range g.Nodes {
		label := node.Title
		if label == "" {
			label = id
		}
		// 节点样式根据类型变化
		sb.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", id, label))
		for _, link := range node.Links {
			sb.WriteString(fmt.Sprintf("  %s --> %s\n", id, link))
		}
		for _, parent := range node.Parents {
			sb.WriteString(fmt.Sprintf("  %s --- %s\n", parent, id))
		}
	}
	sb.WriteString("```\n\n")

	// 生成详细内容
	for _, node := range g.Nodes {
		sb.WriteString("## " + node.Title + " (" + node.ID + ")\n")
		if node.Type != "" {
			sb.WriteString("> Type: " + node.Type + "\n\n")
		}
		sb.WriteString(node.Content + "\n\n")
	}

	return sb.String()
}

// Save 将文档保存为 JSON 文件。
func (g *Graph) Save(filename ...string) error {
	path := g.filename
	if len(filename) > 0 && filename[0] != "" {
		path = filename[0]
	}
	if path == "" {
		return fmt.Errorf("no filename specified")
	}
	return file.Write(path, g.ToJSON())
}

// OpenGraph 从 JSON 文件加载关系型文档。如果文件不存在，则返回一个新的空白文档。
func OpenGraph(filename string) (*Graph, error) {
	g := NewGraph()
	g.filename = filename
	if file.Exists(filename) {
		if err := file.UnmarshalFile(filename, g); err != nil {
			return nil, err
		}
	}
	return g, nil
}
