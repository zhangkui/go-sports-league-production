package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// stat is a tiny helper to count production Go files and lines.
// Run with: go run ./scripts/stat.go
func main() {
	var files, lines int
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") || strings.Contains(path, "vendor") {
			return nil
		}
		if strings.Contains(path, "scripts") {
			return nil
		}
		files++
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(b), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			lines++
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("生产Go文件数: %d\n生产Go代码行数(非空): %d\n", files, lines)
}
