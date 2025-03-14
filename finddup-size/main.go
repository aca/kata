package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const minSize = 100 * 1024 * 1024 // 100 MB in bytes

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <directory>")
		os.Exit(1)
	}
	root := os.Args[1]

	// Map to store file size to list of file paths
	filesBySize := make(map[int64][]string)

	// Walk the directory tree
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// If there was an error accessing the path, skip it.
		if err != nil {
			return nil
		}

		// Only process files (not directories)
		if !info.IsDir() && info.Size() >= minSize {
			filesBySize[info.Size()] = append(filesBySize[info.Size()], path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", root, err)
		os.Exit(1)
	}

	// Print out groups of files with the same size (more than one file)
	for size, files := range filesBySize {
		if len(files) > 1 {
			fmt.Printf("Size: %d bytes\n", size)
			for _, file := range files {
				fmt.Printf("  %s\n", file)
			}
		}
	}
}
