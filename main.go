package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// StructField represents a single field in a struct.
type StructField struct {
	Name     string // empty for embedded (unnamed) fields
	Exported bool
	Type     string
}

// StructDetails holds information about a struct.
type StructDetails struct {
	Name   string
	Fields []StructField
}

// FuncMap holds the name (with signature) of a function or method.
type FuncMap struct {
	Name string
}

// FileMap contains all mapped items of a single file.
type FileMap struct {
	Imports   []string
	Structs   []StructDetails
	Enums     []string
	Traits    []string // interfaces in Go (we keep the name "Traits" for consistent output)
	Functions []FuncMap
	Methods   []FuncMap
}

// nodeToString returns the Go source representation of an AST node.
func nodeToString(n ast.Node) string {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, token.NewFileSet(), n)
	if err != nil {
		return "<error>"
	}
	return buf.String()
}

// formatArgs converts a slice of *ast.Field (e.g., function parameters) into a string.
func formatArgs(fields []*ast.Field) string {
	var parts []string
	for _, f := range fields {
		names := []string{}
		for _, n := range f.Names {
			names = append(names, n.Name)
		}
		typeStr := nodeToString(f.Type)
		if len(names) > 0 {
			parts = append(parts, strings.Join(names, ", ")+" "+typeStr)
		} else {
			parts = append(parts, typeStr)
		}
	}
	return strings.Join(parts, ", ")
}

// formatVisibility returns "pub " if the name is exported, otherwise "".
func formatVisibility(exported bool) string {
	if exported {
		return "pub "
	}
	return ""
}

// hasIota checks recursively whether an expression contains the identifier "iota".
func hasIota(expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name == "iota"
	case *ast.BinaryExpr:
		return hasIota(e.X) || hasIota(e.Y)
	case *ast.UnaryExpr:
		return hasIota(e.X)
	case *ast.ParenExpr:
		return hasIota(e.X)
	case *ast.CallExpr:
		for _, arg := range e.Args {
			if hasIota(arg) {
				return true
			}
		}
	}
	return false
}

// isExportedType returns true if the string representation of an AST type node
// starts with an uppercase letter (exported).
func isExportedType(typ ast.Expr) bool {
	s := nodeToString(typ)
	if s == "" {
		return false
	}
	r := []rune(s)
	if r[0] >= 'A' && r[0] <= 'Z' {
		return true
	}
	// Handle pointer types: *T
	if len(r) > 1 && r[0] == '*' {
		return r[1] >= 'A' && r[1] <= 'Z'
	}
	return false
}

// processFile parses a Go source file and extracts its map.
func processFile(path string) (*FileMap, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	m := &FileMap{}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			switch d.Tok {
			case token.IMPORT:
				for _, spec := range d.Specs {
					imp := spec.(*ast.ImportSpec)
					pathLit := strings.Trim(imp.Path.Value, `"`)
					// Include alias if present
					if imp.Name != nil {
						m.Imports = append(m.Imports, imp.Name.Name+" "+`"`+pathLit+`"`)
					} else {
						m.Imports = append(m.Imports, `"`+pathLit+`"`)
					}
				}

			case token.CONST:
				// Check for iota usage to detect enum-like constants
				usesIota := false
				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)
					for _, val := range vs.Values {
						if hasIota(val) {
							usesIota = true
							break
						}
					}
					if usesIota {
						break
					}
				}
				if usesIota {
					// Determine a name for the enum
					enumName := "int" // default when untyped
					for _, spec := range d.Specs {
						vs := spec.(*ast.ValueSpec)
						if vs.Type != nil {
							enumName = nodeToString(vs.Type)
							break
						}
						// If no explicit type, try to infer from expression type (limited)
						// but we keep default "int"
					}
					m.Enums = append(m.Enums, enumName)
				}

			case token.TYPE:
				for _, spec := range d.Specs {
					ts := spec.(*ast.TypeSpec)
					switch t := ts.Type.(type) {
					case *ast.StructType:
						var fields []StructField
						for _, f := range t.Fields.List {
							fieldType := nodeToString(f.Type)
							if len(f.Names) == 0 {
								// Embedded field (unnamed)
								exported := isExportedType(f.Type)
								fields = append(fields, StructField{
									Name:     "",
									Exported: exported,
									Type:     fieldType,
								})
							} else {
								for _, n := range f.Names {
									exported := ast.IsExported(n.Name)
									fields = append(fields, StructField{
										Name:     n.Name,
										Exported: exported,
										Type:     fieldType,
									})
								}
							}
						}
						m.Structs = append(m.Structs, StructDetails{
							Name:   ts.Name.Name,
							Fields: fields,
						})

					case *ast.InterfaceType:
						m.Traits = append(m.Traits, ts.Name.Name)
					}
				}
			}

		case *ast.FuncDecl:
			name := d.Name.Name
			var args string
			isMethod := d.Recv != nil
			if isMethod {
				// Combine receiver and other parameters
				var allArgs []*ast.Field
				if d.Recv != nil {
					allArgs = append(allArgs, d.Recv.List...)
				}
				allArgs = append(allArgs, d.Type.Params.List...)
				args = formatArgs(allArgs)

				// Determine receiver type string for method name
				recvType := nodeToString(d.Recv.List[0].Type)
				fullName := recvType + "::" + name + "(" + args + ")"
				m.Methods = append(m.Methods, FuncMap{Name: fullName})
			} else {
				// Free function
				args = formatArgs(d.Type.Params.List)
				fullName := name + "(" + args + ")"
				m.Functions = append(m.Functions, FuncMap{Name: fullName})
			}
		}
	}

	return m, nil
}

// printMap prints the file map in a human-readable format identical to the Rust mapper.
func printMap(path string, m FileMap) {
	if len(m.Imports) == 0 && len(m.Structs) == 0 && len(m.Enums) == 0 &&
		len(m.Traits) == 0 && len(m.Functions) == 0 && len(m.Methods) == 0 {
		return
	}

	fmt.Println("\n📄", path)

	if len(m.Imports) > 0 {
		fmt.Println("  📥 Imports:")
		for _, imp := range m.Imports {
			fmt.Println("     -", imp)
		}
	}

	if len(m.Structs) > 0 {
		fmt.Println("  📦 Structs:")
		for _, s := range m.Structs {
			var fieldsDesc string
			if len(s.Fields) > 0 {
				var parts []string
				// Determine if any field is named (we use braces) or all are unnamed (use parens)
				hasNamed := false
				for _, f := range s.Fields {
					if f.Name != "" {
						hasNamed = true
						break
					}
				}
				for _, f := range s.Fields {
					vis := formatVisibility(f.Exported)
					if f.Name != "" {
						parts = append(parts, fmt.Sprintf("%s%s: %s", vis, f.Name, f.Type))
					} else {
						parts = append(parts, fmt.Sprintf("%s%s", vis, f.Type))
					}
				}
				body := strings.Join(parts, ", ")
				if hasNamed {
					fieldsDesc = "{ " + body + " }"
				} else {
					fieldsDesc = "(" + body + ")"
				}
			}
			fmt.Println("     -", s.Name, fieldsDesc)
		}
	}

	if len(m.Enums) > 0 {
		fmt.Println("  🎲 Enums:")
		for _, e := range m.Enums {
			fmt.Println("     -", e)
		}
	}

	if len(m.Traits) > 0 {
		fmt.Println("  📜 Traits:")
		for _, t := range m.Traits {
			fmt.Println("     -", t)
		}
	}

	if len(m.Functions) > 0 {
		fmt.Println("  ⚡ Free Functions:")
		for _, f := range m.Functions {
			fmt.Println("     -", f.Name)
		}
	}

	if len(m.Methods) > 0 {
		fmt.Println("  🔧 Impl Methods:")
		for _, m := range m.Methods {
			fmt.Println("     -", m.Name)
		}
	}
}

func main() {
	targetDir := "."
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Directory '%s' not found.\n", targetDir)
		return
	}

	fmt.Println("🗺️  Generating code map for:", targetDir)

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			m, err := processFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", path, err)
				return nil
			}
			printMap(path, *m)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
	}
}
