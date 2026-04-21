package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileInfo struct {
	Path string
	Size int64
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file1> <file2> ... <fileN>\n", os.Args[0])
		os.Exit(1)
	}

	files := os.Args[1:]
	if len(files) < 2 {
		fmt.Println("Less than 2 files provided. Need at least 2 files.")
		return
	}

	fileInfos, err := getFileInfos(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].Size < fileInfos[j].Size
	})

	fmt.Println("Files sorted by size (smallest to largest):")
	fmt.Println("==========================================")
	for i, fi := range fileInfos {
		fmt.Printf("%d. %s (%d bytes)\n", i+1, fi.Path, fi.Size)
	}
	fmt.Println()

	filesToRemove := fileInfos[:len(fileInfos)-1]
	largestFile := fileInfos[len(fileInfos)-1]

	fmt.Printf("Largest file to keep: %s (%d bytes)\n", largestFile.Path, largestFile.Size)
	fmt.Println("\nFiles to remove:")
	fmt.Println("================")
	for i, fi := range filesToRemove {
		fmt.Printf("%d. %s (%d bytes)\n", i+1, fi.Path, fi.Size)
	}

	if !confirmRemoval() {
		fmt.Println("Operation cancelled.")
		return
	}

	for _, fi := range filesToRemove {
		if err := removeFileInteractive(fi); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", fi.Path, err)
		}
	}

	fmt.Println("\nOperation completed.")
}

func getFileInfos(paths []string) ([]FileInfo, error) {
	var fileInfos []FileInfo
	
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}

		stat, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", absPath, err)
		}

		if stat.IsDir() {
			return nil, fmt.Errorf("%s is a directory, not a file", absPath)
		}

		fileInfos = append(fileInfos, FileInfo{
			Path: absPath,
			Size: stat.Size(),
		})
	}

	return fileInfos, nil
}

func confirmRemoval() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nDo you want to proceed with removing these files? (yes/no): ")
	
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}

func removeFileInteractive(fi FileInfo) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nRemove %s (%d bytes)? (yes/no/skip): ", fi.Path, fi.Size)
	
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	
	switch response {
	case "yes", "y":
		if err := os.Remove(fi.Path); err != nil {
			return err
		}
		fmt.Printf("Removed: %s\n", fi.Path)
	case "skip", "s":
		fmt.Printf("Skipped: %s\n", fi.Path)
	default:
		fmt.Printf("Cancelled removal of: %s\n", fi.Path)
	}

	return nil
}