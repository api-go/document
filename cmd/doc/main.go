package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"apigo.cc/go/cast"
	"apigo.cc/go/document"
)

var (
	jsonOut    = flag.Bool("json", false, "以 JSON 格式输出文档内容")
	mdOut      = flag.Bool("md", false, "以 Markdown 格式输出文档内容 (默认模式)")
	savePath   = flag.String("o", "", "保存结果到指定文件路径 (如: output.xlsx, content.md)")
	createType = flag.String("create", "", "创建新文档，支持类型: xlsx, csv, graph, md")
	password   = flag.String("password", "", "访问加密文档所需的密码 (主要针对 Excel)")
	sheetName  = flag.String("sheet", "", "操作 Excel 时指定的工作表名称或索引 (0, 1...)")
	dataStr    = flag.String("data", "", "注入数据的 JSON 字符串 (支持对象数组或单个对象)")
	inspect    = flag.Bool("inspect", false, "只查看文档元数据 (如类型、页数、工作表列表等)")
	version    = flag.Bool("v", false, "显示版本信息")
)

const docVersion = "1.0.0"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "🗂️  Document CLI (doc) - 极简办公文档处理工具 v%s\n\n", docVersion)
		fmt.Fprintf(os.Stderr, "用法:\n")
		fmt.Fprintf(os.Stderr, "  doc [flags] [file]          # 处理已有文件\n")
		fmt.Fprintf(os.Stderr, "  doc --create [type] [flags]  # 创建新文档\n\n")

		fmt.Fprintf(os.Stderr, "常见示例:\n")
		fmt.Fprintf(os.Stderr, "  doc report.xlsx             # 预览 Excel 内容 (Markdown 表格)\n")
		fmt.Fprintf(os.Stderr, "  doc manual.docx --json      # 提取 Word 内容为结构化 JSON\n")
		fmt.Fprintf(os.Stderr, "  doc paper.pdf -o text.md    # 提取 PDF 文字并存为 Markdown\n")
		fmt.Fprintf(os.Stderr, "  doc --create xlsx -o n.xlsx # 创建空白 Excel\n")
		fmt.Fprintf(os.Stderr, "  doc test.xlsx --data '[{\"ID\":1}]' -o test.xlsx  # 向 Excel 追加数据\n\n")

		fmt.Fprintf(os.Stderr, "参数详解:\n")
		flag.PrintDefaults()
		
		fmt.Fprintf(os.Stderr, "\n支持的格式:\n")
		fmt.Fprintf(os.Stderr, "  Excel (.xlsx), Word (.docx), PDF (.pdf), PPT (.pptx), CSV (.csv), Graph (.graph), Markdown (.md)\n")
	}

	flag.Parse()

	if *version {
		fmt.Printf("doc version %s\n", docVersion)
		return
	}

	args := flag.Args()
	var doc document.Document
	var err error

	// 1. 获取文档实例
	if *createType != "" {
		doc, err = document.Create(*createType)
		if err != nil {
			fail("创建文档失败: %v", err)
		}
	} else if len(args) > 0 {
		filename := args[0]
		if *password != "" {
			doc, err = document.Open(filename, *password)
		} else {
			doc, err = document.Open(filename)
		}
		if err != nil {
			fail("无法打开文件 '%s': %v", filename, err)
		}
	} else {
		flag.Usage()
		return
	}

	// 2. 数据注入逻辑
	if *dataStr != "" {
		applyData(doc, *dataStr, *sheetName)
	}

	// 3. 执行核心操作
	if *inspect {
		runInspect(doc)
		return
	}

	if *savePath != "" {
		if err := doc.Save(*savePath); err != nil {
			fail("保存失败: %v", err)
		}
		fmt.Printf("✨ 成功保存至: %s\n", *savePath)
	} else {
		outputContent(doc, *jsonOut)
	}
}

func applyData(doc document.Document, dataStr, sheet string) {
	var data []map[string]any
	if err := cast.UnmarshalJSON(dataStr, &data); err != nil {
		var single map[string]any
		if err2 := cast.UnmarshalJSON(dataStr, &single); err2 == nil {
			data = []map[string]any{single}
		} else {
			fail("数据格式无效，请提供有效的 JSON 对象或数组: %v", err)
		}
	}

	switch d := doc.(type) {
	case *document.Excel:
		if err := d.SetData(sheet, data, "A1", ""); err != nil {
			fail("写入 Excel 失败: %v", err)
		}
	case *document.Graph:
		fmt.Println("⚠️  提示: Graph 类型目前主要通过 API 操作，暂不支持通过 CLI 批量 SetData。")
	default:
		fmt.Printf("⚠️  警告: 当前文档类型 (%T) 不支持数据注入操作。\n", d)
	}
}

func runInspect(doc document.Document) {
	fmt.Printf("🔍 文档详情:\n")
	fmt.Printf("  类型: %T\n", doc)
	
	switch d := doc.(type) {
	case *document.Excel:
		fmt.Printf("  工作表: %s\n", strings.Join(d.Sheets(), ", "))
	case *document.PDF:
		if pages, ok := d.Metadata["pages"]; ok {
			fmt.Printf("  总页数: %v\n", pages)
		}
		for k, v := range d.Metadata {
			if k != "pages" {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
	}
}

func outputContent(doc document.Document, asJSON bool) {
	if asJSON {
		fmt.Println(doc.ToJSON())
	} else {
		content := doc.ToMarkdown()
		if content == "" {
			fmt.Println("(文档内容为空)")
		} else {
			fmt.Println(content)
		}
	}
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "❌ 错误: "+format+"\n", a...)
	os.Exit(1)
}
