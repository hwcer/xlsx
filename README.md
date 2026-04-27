# xlsx

> ⚠️ **警告:本项目由 AI 接管维护,不建议碳基生物阅读代码。**
> 如需修改或排障,请通过 AI 代理进行;人类直读源码可能引发困惑、血压升高及不可逆的颈椎损伤。

Excel 打表工具:将 Excel 配置表批量转换为 `.proto` 定义、JSON 数据以及可选的多语言文件。

## 特性

- 读取目录下所有 `.xlsx` 配置表,统一生成一份 `.proto` 文件
- 支持 `map` / `struct(kv)` 两种 Sheet 类型
- 支持嵌套对象 (`Dummy`) 与自动去重的全局对象声明
- 支持总表 (Summary) 生成
- 支持从源表的指定列衍生枚举表 (Enum)
- 支持使用现有 `.proto` 文件作为输出头部,替代内置模板
- 支持多语言文本抽取
- 支持自定义 Parser、输出插件、命名过滤器

## 安装 / 构建

以 `example/` 为参考:

```bat
SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -o ./bin/xlsx.exe .
```

入口只需注册 Parser 并启动 cosgo:

```go
package main

import (
    "github.com/hwcer/cosgo"
    "github.com/hwcer/xlsx"
    "github.com/hwcer/xlsx/sample"
)

func init() {
    xlsx.Config.Parser = sample.New
}

func main() {
    cosgo.Start(false, xlsx.New())
}
```

## 命令行参数

| 参数         | 说明                                          |
|--------------|-----------------------------------------------|
| `--in`       | Excel 源目录(或单个文件)                     |
| `--out`      | 输出目录(生成 `.proto`)                      |
| `--go`       | proto → Go 代码的输出目录                     |
| `--json`     | 导出 JSON 数据的目录                          |
| `--tag`      | 字段标记,区分前后端,如 `C` / `S`(默认 `S`) |
| `--ignore`   | 忽略的文件或目录,逗号分隔                    |
| `--branch`   | 使用特定版本分支                              |
| `--package`  | proto `package` 名称(默认 `protoc`)          |
| `--GOPackage`| `option go_package`(默认跟随 `package`)      |
| `--CSPackage`| `option csharp_namespace`(默认跟随 `package`)|
| `--summary`  | 总表名称,留空不生成                          |
| `--language` | 多语言 Excel 输出路径                         |
| `--verify`   | 是否开启空值检查                              |

示例 (`example/run.bat`):

```bat
.\bin\xlsx --ignore="MsgList,client" --in="./excel" --out="./output" --go="./output" --json="./data"
```

## 配置文件 `config.toml`

CLI 参数(小写键,如 `in`/`out`/`go`/`json`/`tag` 等)与 `xlsx.Config` 结构体字段(大写键,如 `Package`/`GOPackage`/`ProtoHeader` 等)均可写入 `config.toml`。前者由 cosgo flag 读取,后者通过 `Unmarshal(Config)` 注入:

```toml
in      = "./excel"
out     = "./output"
json    = "./output/data/"
ignore  = ""
branch  = ""

Package     = "protoc"
GOPackage   = "configs"
CSPackage   = "Cosnet.configs"
Proto       = "configs.proto"   # 输出 proto 文件名
Summary     = ""                # 总表名,留空不生成
ProtoHeader = ""                # 可选,指向现有 proto 文件,用其内容替代内置文件头部模板 (设置后 Package/GOPackage/CSPackage 失效)
```

### 自定义 proto 文件头部 `ProtoHeader`

默认情况下 proto 文件头由内置模板生成:

```
syntax = "proto3";
option go_package = "./;<GOPackage>";
option csharp_namespace = "<CSPackage>";
package <Package>;
```

若设置了 `ProtoHeader`,工具会直接读取该文件内容写入输出 proto 文件的头部(不做模板替换)。

> ⚠️ **此时 `Package` / `GOPackage` / `CSPackage` 配置均失效**,`syntax`、`option go_package`、`option csharp_namespace`、`package` 等声明需要用户自行在引用的 proto 文件里完整书写。

```toml
ProtoHeader = "./header.proto"
```

#### 配合 `NamedDummyInHeader` 避免子对象重复

若 Excel 中用 `<Name>` / `field.Name{...}` 等语法显式命名了子对象,且这些对象已在 `ProtoHeader` 文件中声明过,可启用该选项跳过它们的 `message` 生成,避免与头文件重复定义导致 protoc 编译失败:

```toml
ProtoHeader           = "./base.proto"
EnableGlobalDummyName = false       # 禁止自动生成子对象名称,必须通过 field.Name{} 显式指定
NamedDummyInHeader    = true        # 命名子对象假定已在 base.proto 中声明,跳过生成
```

- 显式命名的子对象(通过 `<>` 或 `.Name` 指定)不会注册到全局对象,视为已在 `ProtoHeader` 中定义
- 匿名子对象(未指定名字的嵌套对象)仍会按签名自动命名并写入输出 proto(`EnableGlobalDummyName` 须为 `true`)

### 枚举配置 `[enum]`

允许基于现有源表的若干列,衍生出一张新的枚举 Sheet。配置格式:

```toml
[enum.ItemType]    # key = 新生成的枚举名
Src   = "Item"     # 源表 ProtoName
Index = [0, 1, 2, 3]  # [key列, val列, type列, desc列]
```

- `Src`: 源表 ProtoName,工具在导表时匹配该表并附加枚举生成规则
- `Index`:
  - `[0]` key 列:枚举项名
  - `[1]` val 列:枚举项对应值(即 `ProtoIndex`)
  - `[2]` type 列:proto 类型,留空/设为 `-1` 时默认为 `int32`
  - `[3]` desc 列:注释,设为 `-1` 时省略

工具会把枚举名通过 `TrimProtoName` 进行 CamelCase 归一化(`Src` 同理),配置里大小写与下划线不敏感。

也可以在代码里注册:

```go
xlsx.Config.SetEnum("ItemType", "Item", [4]int{0, 1, 2, 3})
```

另一种方式:在源表 Sheet 首行通过 Parser 约定触发(完整格式见 [首行语法](#首行语法-第-1-行)),例如首行任一单元格填 `kv:ItemType:0,1,2,3` 会自动调用 `sheet.AddEnum`。

## 编程式 API

`xlsx.Config` 暴露的常用入口:

| 方法                                | 说明                                        |
|-------------------------------------|---------------------------------------------|
| `SetType(t SheetType, names ...)`   | 注册 Sheet 类型别名(默认已包含 map/hash/kv 等)|
| `GetType(name) SheetType`           | 读取 Sheet 类型                              |
| `SetEnum(name, src, index [4]int)`  | 注册枚举生成规则                             |
| `SetOutput(o Output)`               | 附加自定义输出插件(实现 `Writer([]*Sheet)`) |
| `SetJsonNameFilter(f)`              | 自定义 JSON 文件命名                         |
| `SetProtoNameFilter(f)`             | 自定义 proto message 命名                    |

其它可赋值字段:

- `Config.Parser func(*Sheet) Parser`:**必填**,Sheet 解析器工厂
- `Config.Empty func(string) bool`:自定义空值判定(默认空字符串)
- `Config.Message func() string`:向 proto 注入额外全局对象
- `Config.Language []string`:被视为多语言文本的 `FieldType`(默认 `text`/`lang`/`language`)
- `Config.LanguageNewSheetName`:多语言增量页签名(默认 `多语言文本`)
- `Config.EnableGlobalDummyName bool`:是否允许未显式命名的子对象按签名自动生成名称;为 `false` 时所有子对象必须通过 `.Name{}`/`<Name>` 显式命名
- `Config.NamedDummyInHeader bool`:显式命名的子对象假定已在 `ProtoHeader` 中声明,不注册到全局对象也不生成 `message`(避免与头文件重复)

## Parser 约定

Sheet 解析交由 `Parser` 接口完成:

```go
type Parser interface {
    Verify() (skip int, name string, ok bool) // 校验 Sheet 并返回要跳过的行数、Sheet 名
    Fields() []*Field                         // 返回字段定义
}
```

`sample/` 包提供了一份参考实现,约定:

- 第 1 行:A 列为数据表名,其它单元格为附加配置(见下表)
- 第 2 行:proto 类型关键字(见下表)
- 第 3 行:字段名 / 嵌套结构声明(见下表,可带 `#branch` 版本分支)
- 第 4 行:字段描述

#### 首行语法 (第 1 行)

| 单元格   | 规则                                                                                      |
|----------|-------------------------------------------------------------------------------------------|
| A1       | 数据表名(必填),将经 `TrimProtoName` 归一化为 proto message 名                           |
| B1…Z1    | 可选附加配置;目前唯一支持的格式为 `kv` 模式,用于基于本表派生一张枚举 Sheet              |

**`kv` 模式 —— 从本表列派生枚举**

附加单元格写入 `kv[:EnumName:idx_list]` 即可触发,由 `sample/parser.go:Verify` 调用 `sheet.AddEnum`,随后 `reParseEnum` 逐行读取指定列生成新 Sheet:

| 写法                            | 效果                                                                                  |
|---------------------------------|---------------------------------------------------------------------------------------|
| `kv`                            | 默认:以当前表名为枚举名,列索引 `[0,1,2,3]`(key, val, type, desc)                  |
| `kv:EnumName:k,v`               | 自定义枚举名,指定 key/val 两列,type 默认 `int32`,desc 省略                          |
| `kv:EnumName:k,v,t`             | 同上,并指定 type 列                                                                  |
| `kv:EnumName:k,v,t,d`           | 完整指定 4 列;某列填 `-1` 表示省略(类型会回落为 `int32`,desc 为空)                |

- `EnumName` 会经 `TrimProtoName` 归一化(首字母大写、下划线转驼峰)
- 索引从 0 开始,依次为 **key 列**(枚举项名)、**val 列**(值/`ProtoIndex`)、**type 列**(proto 类型,默认 `int32`)、**desc 列**(注释)
- 同一张表可以写多个 `kv:...` 单元格,派生多张枚举

示例(`example/base.proto` 思路):A1=`Base`,B1=`kv:GameConfig:0,1`。这样在生成时除了原表 `Base` 外,还会得到一张新 Sheet `GameConfig`,把 Base 的第 1 列作 key、第 2 列作 val 组装成 KV 表。

> 代码里也可以用 `xlsx.Config.SetEnum("GameConfig", "Base", [4]int{0,1,-1,-1})` 达到等价效果;或者通过 `config.toml` 里的 `[enum]` 段配置(见上文 **枚举配置 `[enum]`** 章节)。三种方式可共存。

#### 类型关键字 (第 2 行)

| 关键字                                   | 含义                                          |
|------------------------------------------|-----------------------------------------------|
| `int` / `int32` / `num` / `number`       | int32                                         |
| `int64`                                  | int64                                         |
| `uint32` / `uint64`                      | uint32 / uint64                               |
| `float` / `float32`                      | float                                         |
| `float64` / `double`                     | double                                        |
| `bool` / `byte`                          | bool / bytes                                  |
| `str` / `string` / `text` / `lang` / `language` | string(多语言类型会被收入语言包)       |
| `Object`                                 | 单个子对象                                    |
| `ArrObject`                              | 子对象数组                                    |
| `ArrInt` / `ArrInt32` / `ArrInt64` / `ArrFloat` / `ArrFloat64` / `ArrString` | **多单元格**填充的基础类型数组 |
| `[]int` / `[]int32` / `[]int64` / `[]float` / `[]float64` / `[]string`       | **单单元格**逗号分隔的基础类型数组 |

#### 字段名单元格语法 (第 3 行)

| 语法                      | 含义                                                     |
|---------------------------|----------------------------------------------------------|
| `name`                    | 普通字段(对应基础类型或 `[]xxx` 切分数组)              |
| `name[` ... `]`           | 多单元格数组:`name[` 起始单元格,数据按列顺序填充,`]` 写在最后一个单元格 |
| `name{` ... `}`           | 多单元格对象:`name{` 起始,子字段按列填充,`}` 闭合       |
| `name[{` ... `}]`         | 多单元格数组对象:`name[{` 起始,`}]` 闭合                |
| `<alias>...`              | 把字段名重写为 `alias`(同时作为子结构体名)             |
| `field.dummy{...}`        | 对象字段,显式指定子结构体名为 `dummy`                   |
| `field.dummy[{...}]`      | 数组对象字段,显式指定子结构体名为 `dummy`               |
| `name#branch`             | 版本分支字段,不同 branch 共享同一 proto 字段位置        |

显式指定子结构体名时,`.` 之后的部分会经 `TrimProtoName` 格式化为 PROTO 命名规范(首字母大写、剥离下划线驼峰化)。若未显式指定,子结构体名会根据字段签名自动生成,可配合 `Config.EnableGlobalDummyName` 控制是否保留自定义名。

实际项目建议直接基于 `sample` 扩展或编写自己的 Parser。

## 输出内容

- `out/<Proto>`:合并后的 proto 文件(全局对象、数据对象、可选的总表)
- `json/`:每张 Sheet 一个 JSON 文件
- `language/`:多语言 Excel(启用时)
- `go/`:proto 对应的 Go 代码(启用时)

## 示例

见 `example/`:

- `example/excel/` 原始 Excel
- `example/config.toml` 配置
- `example/output/configs.proto` 生成结果
- `example/output/data/*.json` JSON 数据

示例引入了 `sample/info.go`,它额外注册了一个 `info` flag(`--info` 或 config.toml 里的 `info = "..."`),用于输出一份 JSON 索引(Sheet 名 → 类型/文件路径/主键类型)。这不是核心功能,使用 `sample` 包时才会生效,自定义 Parser 可以不导入。
