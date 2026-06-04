# AST Go Mapper

A lightweight command-line tool that parses Go source files and prints a human-readable **code map** of the package’s public API.  
It identifies imports, structs (with fields and visibility), enums (constants that use `iota`), interfaces, free functions, and methods, presenting them in a tree-like format similar to a Rust module outline.

## Features

- 📥 **Imports** – lists all import statements, including aliases.
- 📦 **Structs** – shows struct name, field names, types, and export/visibility (prefixed with `pub` when exported).  
  Unnamed (embedded) fields are displayed without a name.
- 🎲 **Enums** – detects type-safe `iota` constant blocks and displays the base type (e.g., `int`).
- 📜 **Traits** – surfaces all `interface` types (named "Traits" for consistency).
- ⚡ **Free Functions** – standalone functions with their full signature.
- 🔧 **Impl Methods** – methods shown as `ReceiverType::methodName(params)`.

## Installation

Requires **Go 1.26.3** or later (the version specified in `go.mod`).

### Option 1: Install directly with `go install`

```bash
go install github.com/rodolfovillaruz/ast-go-mapper@latest
```

This downloads, compiles, and places the binary in your `$GOPATH/bin` (or `$GOBIN`). Make sure that directory is in your system’s `PATH`.

### Option 2: Build from source

```bash
git clone https://github.com/rodolfovillaruz/ast-go-mapper.git
cd ast-go-mapper
go build -o ast-go-mapper .
```

Then move the `ast-go-mapper` binary to a location in your `PATH`, or run it directly with `./ast-go-mapper`.

## Usage

Run the tool from the root of your Go project (or any directory containing `.go` files):

```bash
ast-go-mapper
```

### Example output

```
🗺️  Generating code map for: .

📄 func.go
  📥 Imports:
     - "bytes"
     - "fmt"
     - "go/ast"
     - "go/parser"
     - "go/printer"
     - "go/token"
     - "strings"
  ⚡ Free Functions:
     - nodeToString(n ast.Node)
     - formatArgs(fields []*ast.Field)
     - formatVisibility(exported bool)
     - hasIota(expr ast.Expr)
     - isExportedType(typ ast.Expr)
     - processFile(path string)
     - printMap(path string, m FileMap)

📄 main.go
  📥 Imports:
     - "fmt"
     - "os"
     - "path/filepath"
     - "strings"
  ⚡ Free Functions:
     - main()

📄 struct.go
  📦 Structs:
     - StructField { pub Name: string, pub Exported: bool, pub Type: string }
     - StructDetails { pub Name: string, pub Fields: []StructField }
     - FuncMap { pub Name: string }
     - FileMap { pub Imports: []string, pub Structs: []StructDetails, pub Enums: []string, pub Traits: []string, pub Functions: []FuncMap, pub Methods: []FuncMap }
```

## How it works

The program uses Go’s standard `go/ast`, `go/parser`, and `go/printer` packages to parse each `.go` file into an abstract syntax tree. It then traverses top-level declarations, collecting:

- Import specs
- Type declarations (structs, interfaces)
- Constant blocks that use `iota`
- Function and method declarations with their full signatures

The collected data is printed to stdout in a human‑friendly, emoji‑annotated format.

## License

This project is provided as-is without a specific license. Feel free to use and modify it for your own needs.