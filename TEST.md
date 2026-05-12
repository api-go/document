# document TEST

## 覆盖场景
- [x] **Excel 基础读写**: 验证 `NewExcel`, `Set`, `Get`, `Save`, `OpenExcel`。
- [x] **Excel 对象映射**: 验证 `SetData`, `GetData` 及动态列扩展。
- [x] **统一 API**: 验证 `document.Open` 自动识别、`ToJSON` 和 `ToMarkdown`。
- [x] **ID 生成解析**: 验证 `MakeCellID` 和 `ParseCellID` 的准确性。

## 性能测试 (Benchmark)
- **环境**: Darwin / Intel(R) Core(TM) i7-1068NG7 CPU @ 2.30GHz
- **Excel.SetData**: 写入 100 行对象数据。

```text
BenchmarkExcel_SetData-8             712           1673885 ns/op
```

## 基础设施对齐
- 全面使用 `apigo.cc/go/cast` 替代原生类型转换。
- 移除了所有原生 `os` 和 `strconv` 依赖。
