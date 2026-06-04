package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
