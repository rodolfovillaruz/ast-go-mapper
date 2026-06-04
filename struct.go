package main

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
