# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Excel-to-Protocol Buffer code generation tool (`github.com/hwcer/xlsx`). Reads `.xlsx` / `.csv` configuration tables and generates `.proto` definitions, JSON data files, and optionally Go code via `protoc`. Written in Go 1.25, uses the `cosgo` framework for CLI/config lifecycle.

## Build & Run

```bat
SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -o ./bin/xlsx.exe ./example/
```

Run with flags or `config.toml`:
```bat
.\bin\xlsx --in="./excel" --out="./output" --go="./output" --json="./data"
```

There are no tests in this project. Use `go vet ./...` for static checks.

## Architecture

### Lifecycle (cosgo module)

`module.go` registers CLI flags in `init()` and wires the main flow in `Module.Start()`:
1. `cosgo.Config.Unmarshal(Config)` — loads config.toml into the global `Config` struct
2. `cosgo.Config.UnmarshalKey("enum", ...)` — loads `[enum.*]` sections
3. `preparePath()` — resolves output directories
4. `LoadExcel(inputDir)` — the main pipeline

### Data Pipeline (`excel.go` → output files)

```
LoadExcel(dir)
  → GetFiles() filters by Config.Extensions (.xlsx, .csv)
  → OpenFile() per file (csv.go: CSV → excelize.File, otherwise excelize.OpenFile)
  → parseSheet() per sheet:
      Sheet created → Config.Parser(sheet) → Parser.Verify() + Parser.Fields()
      → builds Field list, resolves Dummy (nested objects), attaches enums
  → writeProtoMessage() → .proto file
  → writeValueJson() → JSON data files
  → ProtoGo() → shells out to `protoc` for Go code
  → writeLanguage() → multi-language Excel (if configured)
  → Config.Outputs[].Writer() → custom plugins
```

### Key Types

- **`Sheet`** (`model_sheet.go`): Represents one parsed Excel sheet. Has `Fields`, `ProtoName`, `SheetType` (Hash or Enum/KV). Produces row data via `Values()`.
- **`Field`** (`model_fields.go`): One column/field. Holds `ProtoType`, optional `Dummy` children for nested objects, optional `Branch` variants for versioned fields.
- **`Dummy`** (`model_dummy.go`): Nested proto message (sub-object). Fields compiled from bracket syntax in Excel column headers. Deduplicated globally by signature.
- **`Parser`** interface (`config.go`): `Verify() (skip, name, ok)` + `Fields() []*Field`. Factory registered via `Config.Parser`.
- **`config`** (`config.go`): Global singleton (`xlsx.Config`). All customization goes through this struct — parser factory, output plugins, naming filters, type registry.

### Type System (`types.go` + `sample/types.go`)

Proto types are registered via `Register(ProtoBuffType, ProtoBuffParse)`. The `ProtoBuffParse` interface provides `Type() string`, `Value(...string) (any, error)`, and `Repeated() bool`. The `sample` package registers extended types: `Object`, `ArrObject`, `ArrInt`, `[]int`, etc. Custom types can be added by implementing `ProtoBuffParse` and calling `Register()`.

### Parser Convention (`sample/` package)

The reference parser in `sample/` expects a 4-row header format:
- Row 1: Table name (A1) + optional `kv:EnumName:col_indices` attachments
- Row 2: Proto type keywords (`int32`, `string`, `Object`, `ArrInt`, `[]int`, etc.)
- Row 3: Field names with structure syntax (`name`, `name[...]`, `name{...}`, `name[{...}]`, `field.DummyName{...}`, `<DummyName>field{...}`, `name#branch`)
- Row 4: Field descriptions

Fields and table names support `S:`/`C:` prefixes for server/client tagging, filtered by `--tag`.

### CSV 支持 (`csv.go`)

`OpenFile(file)` 是统一的文件打开入口，根据扩展名分发：`.csv` 通过 `openCSV` 读取后写入 `excelize.File`（sheet 名取自文件名），其余走 `excelize.OpenFile`。业务层无需感知格式差异，CSV 文件遵循与 Excel 相同的 4 行头部约定。

### 关键配置项

- **`Extensions`**: 有效输入文件扩展名，默认 `[".xlsx", ".csv"]`，`Ignore()` 过滤时使用
- **`ArraySplitString`**: 单元格切割数组（`[]int` 等）的分隔符优先级列表，默认 `[",", "-", "_", "|", ";", ":"]`，匹配第一个出现在单元格内容中的分隔符

### Output Plugin System

Implement `Output` interface (`Writer(sheets []*Sheet)`) and register via `Config.SetOutput(o)`. The `sample/info.go` demonstrates this — it generates a JSON index of all sheets.

## Language

All user-facing strings, comments, log messages, and documentation are in Chinese. Maintain this convention.