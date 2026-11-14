# Test Report

## v1.5.9 验证
- `TestDocxMarkdownIncludesEmbeddedImage` 覆盖 DrawingML/VML 图片关系解析、轻量 `doc-asset://` 引用与资产二进制提取。
- `go test -v ./...` 与 `go test -bench=. ./...` 均通过。

## v1.5.7 验证
- `go list -m all`、`go test ./...`、`go test -bench=. ./...` 均通过。
- Markdown 测试文件名不再与仓库 `TEST.md` 在大小写不敏感文件系统上冲突。

## GitHub migration validation (pending release)

Module paths and internal dependency versions are migrated. Development verification uses GOWORK=off and explicit sibling replacements; remove these replacements before release. See migration validation logs for actual results. Existing unrelated working changes are preserved and require separate review.

API inventory verification: `GOWORK=off go doc` succeeded; output retained in the migration validation directory. Source exports were preserved by the path migration.

## GitHub release verification

GOWORK=off go test -v ./... and go test -bench=. ./... passed using published GitHub dependencies, without local replacements. go mod verify passed.

## 依赖联动验证（2026-11-14，未发布）

GOWORK=off，显式替换 safe/crypto；test: 0, bench: 0。0 表示通过；非零或超时结果待定位，不能视为验收通过。

## 最终依赖图验证（2026-11-14，未发布）

显式本地 replace 覆盖本轮待发布依赖链；GOWORK=off go mod tidy、go test -v ./...、go test -bench=. -benchtime=1x ./...。基准按固定 1 次采样，避免排除计时的昂贵准备阶段反复校准。本轮结果：[('tidy', '0'), ('test', '0'), ('bench', '0')]。0 为通过，其他结果待定位。

固定一次采样的基准摘要（非稳定性能测量）：

```text
BenchmarkExcel_SetData-16    	       1	    203750 ns/op
```

## v1.5.11 正式发布验证（2026-11-14）

GOWORK=off；移除全部 replace，使用已发布依赖。go mod tidy、go test -v ./...、go test -bench=. ./...、go mod verify 均通过。API 工具不可用，使用 go doc 与源码核对。

```text
BenchmarkExcel_SetData-16    	    1405	    932319 ns/op
```
