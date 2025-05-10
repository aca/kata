package main

import (
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/cespare/xxhash/v2"
)

var (
	minSizeMB = flag.Int("min-size", 0, "Minimum file size in MB to consider")
	useCRC32  = flag.Bool("crc32", false, "Use CRC32 instead of xxhash")
	doDelete  = flag.Bool("delete", false, "Actually delete duplicates")
)

// getFileHash computes the hash of a file using xxhash or CRC32
func getFileHash(path string, useCRC32 bool) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if useCRC32 {
		h := crc32.NewIEEE()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
		return fmt.Sprintf("%08x", h.Sum32()), nil
	}

	h := xxhash.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum64()), nil
}

// findDuplicates 그룹별 해시를 계산하여 중복 파일 그룹을 반환
func findDuplicates(root string, minBytes int64) ([][]string, error) {
	sizeMap := make(map[int64][]string)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		size := info.Size()
		if size >= minBytes {
			sizeMap[size] = append(sizeMap[size], path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var duplicates [][]string
	for _, paths := range sizeMap {
		if len(paths) < 2 {
			continue
		}
		hashMap := make(map[string][]string)
		for _, p := range paths {
			h, err := getFileHash(p, *useCRC32)
			if err != nil {
				log.Printf("Error hashing %s: %v", p, err)
				continue
			}
			hashMap[h] = append(hashMap[h], p)
		}
		for _, group := range hashMap {
			if len(group) > 1 {
				duplicates = append(duplicates, group)
			}
		}
	}
	return duplicates, nil
}

// removeDuplicates 중복 파일을 실제 삭제 또는 드라이런
func removeDuplicates(groups [][]string, dryRun bool) {
	for _, group := range groups {
		sort.Strings(group)
		for _, dup := range group[1:] {
			if dryRun {
				fmt.Printf("rm -v %q\n", dup)
			} else {
				if err := os.Remove(dup); err != nil {
					fmt.Printf("Error removing %s: %v\n", dup, err)
				} else {
					fmt.Printf("Removed: %s\n", dup)
				}
			}
		}
	}
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	if root == "" {
		log.Fatal("Usage: remove_duplicates.go [options] <directory>")
	}
	minBytes := int64(*minSizeMB) * 1024 * 1024
	fmt.Printf("Scanning %s for files >= %d MB...\n", root, *minSizeMB)
	dups, err := findDuplicates(root, minBytes)
	if err != nil {
		log.Fatalf("Scan error: %v", err)
	}
	if len(dups) == 0 {
		fmt.Println("No duplicates found.")
		return
	}
	total := 0
	for _, g := range dups {
		total += len(g) - 1
	}
	fmt.Printf("Found %d duplicates in %d groups.\n", total, len(dups))
	removeDuplicates(dups, !*doDelete)
}
