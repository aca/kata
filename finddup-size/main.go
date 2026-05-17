package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// shellQuote wraps s in single quotes, escaping any embedded single quotes
// so the result is safe to paste into a POSIX shell.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

const (
	minSize   = 100 * 1024 * 1024 // 100 MB in bytes
	blockSize = 1024
)

// signature returns the concatenation of the first and last blockSize bytes of the file.
func signature(path string, size int64) ([2 * blockSize]byte, error) {
	var sig [2 * blockSize]byte

	f, err := os.Open(path)
	if err != nil {
		return sig, err
	}
	defer f.Close()

	if _, err := io.ReadFull(f, sig[:blockSize]); err != nil {
		return sig, err
	}
	if _, err := f.ReadAt(sig[blockSize:], size-blockSize); err != nil {
		return sig, err
	}
	return sig, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <directory>")
		os.Exit(1)
	}
	root := os.Args[1]

	filesBySize := make(map[int64][]string)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Size() >= minSize {
			filesBySize[info.Size()] = append(filesBySize[info.Size()], path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", root, err)
		os.Exit(1)
	}

	// Cache of directory entry counts so we don't re-read the same dir.
	dirCounts := make(map[string]int)
	siblings := func(path string) int {
		dir := filepath.Dir(path)
		if n, ok := dirCounts[dir]; ok {
			return n
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			dirCounts[dir] = 0
			return 0
		}
		dirCounts[dir] = len(entries)
		return len(entries)
	}

	var savings int64
	for size, files := range filesBySize {
		if len(files) < 2 {
			continue
		}

		// Group by (first 1k + last 1k) signature.
		bySig := make(map[[2 * blockSize]byte][]string)
		for _, f := range files {
			sig, err := signature(f, size)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skip %s: %v\n", f, err)
				continue
			}
			bySig[sig] = append(bySig[sig], f)
		}

		for _, dups := range bySig {
			if len(dups) < 2 {
				continue
			}
			savings += int64(len(dups)-1) * size

			// Keep the file whose parent directory has the most entries.
			// Stable sort so ties keep walk order.
			sort.SliceStable(dups, func(i, j int) bool {
				return siblings(dups[i]) > siblings(dups[j])
			})

			fmt.Printf("# Size: %d bytes\n", size)
			for i, f := range dups {
				abs, err := filepath.Abs(f)
				if err != nil {
					abs = f
				}
				if i == 0 {
					fmt.Printf("# keep %s (siblings=%d)\n", shellQuote(abs), siblings(f))
					continue
				}
				fmt.Printf("rm %s # siblings=%d\n", shellQuote(abs), siblings(f))
			}
		}
	}

	fmt.Printf("# Reclaimable: %s (%d bytes)\n", humanBytes(savings), savings)
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n/div >= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
