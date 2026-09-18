# Test Report

## v1.5.9 验证
- `TestDocxMarkdownIncludesEmbeddedImage` 覆盖 DrawingML/VML 图片关系解析、轻量 `doc-asset://` 引用与资产二进制提取。
- `go test -v ./...` 与 `go test -bench=. ./...` 均通过。

## v1.5.7 验证
- `go list -m all`、`go test ./...`、`go test -bench=. ./...` 均通过。
- Markdown 测试文件名不再与仓库 `TEST.md` 在大小写不敏感文件系统上冲突。
