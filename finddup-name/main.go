package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileDetail holds the file path and its size.
type FileDetail struct {
	Path string
	Size int64
}

func main() {
	// Parse command line arguments.
	dirPtr := flag.String("dir", ".", "Directory to scan for duplicate files")
	rmFlag := flag.Bool("rm", false, "If set, remove files that are smaller than the largest file in the duplicate group")
	flag.Parse()

	// Map to group files by base name (without extension).
	filesMap := make(map[string][]FileDetail)

	// Walk through the directory tree.
	err := filepath.Walk(*dirPtr, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Process only files.
		if !info.IsDir() {
			baseName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			filesMap[baseName] = append(filesMap[baseName], FileDetail{Path: path, Size: info.Size()})
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the directory: %v\n", err)
		os.Exit(1)
	}

	// Process duplicate groups.
	fmt.Println("Duplicate files (files with similar base names):")
	found := false
	for _, group := range filesMap {
		if len(group) > 1 {
			// Determine the maximum file size in the group.
			var maxSize int64 = 0
			for _, file := range group {
				if file.Size > maxSize {
					maxSize = file.Size
				}
			}
			// Process each file in the group.
			for _, file := range group {
				if *rmFlag && file.Size <= maxSize {
					// Remove file if its size is smaller than the max.
					err := os.Remove(file.Path)
					if err != nil {
						fmt.Printf("Failed to remove %s - %d bytes: %v\n", file.Path, file.Size, err)
					} else {
						fmt.Printf("Removed %s - %d bytes\n", file.Path, file.Size)
					}
				} else {
					// Keep the file.
					fmt.Printf("Keeping %s - %d bytes\n", file.Path, file.Size)
				}
			}
			fmt.Println()
			found = true
		}
	}
	if !found {
		fmt.Println("No duplicates found.")
	}
}
