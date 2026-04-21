package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Parse command-line arguments.
	flag.Parse()
	dirs := flag.Args()
	if len(dirs) < 2 {
		fmt.Println("Usage: go run main.go <directory1> <directory2> ...")
		os.Exit(1)
	}

	// Map to hold relative paths and their corresponding full paths.
	duplicates := make(map[string][]string)

	// Walk through each directory.
	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if strings.Contains(path, "snapraid") {
				return nil
			}

			// Skip directories.
			if d.IsDir() {
				return nil
			}
			// Compute the relative path from the root directory.
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			duplicates[rel] = append(duplicates[rel], path)
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory %s: %v\n", dir, err)
		}
	}

	// Report duplicates.
	fmt.Println("Duplicate files (same relative path across different directories):")
	for rel, paths := range duplicates {
		if len(paths) > 1 {
			fmt.Printf("\nRelative Path: %s\n", rel)
			for _, p := range paths {
				fmt.Printf("  %s\n", p)
			}
		}
	}
}
