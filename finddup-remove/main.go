// remove_duplicates.go
// Go 프로그램으로 디렉토리 내 5MB 이상의 중복 파일을 즉시 찾아 제거하되,
// 중복이 확인된 파일에 한해 해시를 계산합니다.
package main

import (
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cespare/xxhash/v2"
)

var (
	minSizeMB = flag.Int("min-size", 5, "Minimum file size in MB to consider")
	useCRC32  = flag.Bool("crc32", false, "Use CRC32 instead of xxhash")
	doDelete  = flag.Bool("delete", false, "Actually delete duplicates")
)

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

func main() {
	flag.Parse()
	root := flag.Arg(0)
	if root == "" {
		log.Fatal("Usage: remove_duplicates.go [options] <directory>")
	}
	minBytes := int64(*minSizeMB) * 1024 * 1024
	fmt.Printf("Scanning %s for files >= %d MB...\n", root, *minSizeMB)

	// size -> first file path (unhashed)
	firstBySize := make(map[int64]string)
	// size -> (hash -> file path), active after first collision
	hashedMap := make(map[int64]map[string]string)
	removedCount := 0

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
		if size < minBytes {
			return nil
		}

		// 첫 파일인지 확인
		if orig, seen := firstBySize[size]; !seen {
			// 아직 같은 크기의 파일을 본 적 없음
			firstBySize[size] = path
			return nil
		} else {
			// 두 번째 이상의 파일
			if _, ok := hashedMap[size]; !ok {
				// 해시 맵이 없으면, 처음 충돌이 발생한 것이므로
				// 첫 파일과 현재 파일 해시 계산
				hashOrig, err1 := getFileHash(orig, *useCRC32)
				hashCur, err2 := getFileHash(path, *useCRC32)
				if err1 != nil || err2 != nil {
					log.Printf("Hash error: %v, %v", err1, err2)
					// 해시 에러 시 두 파일 모두 맵에 저장
					hashedMap[size] = map[string]string{}
					hashedMap[size]["<hash_err>_"+orig] = orig
					hashedMap[size]["<hash_err>_"+path] = path
					return nil
				}
				hashedMap[size] = map[string]string{hashOrig: orig}
				// 현재 파일에 대해 비교
				if hashCur == hashOrig {
					if *doDelete {
						os.Remove(path)
						fmt.Printf("Removed duplicate: %s (original: %s)\n", path, orig)
					} else {
						fmt.Printf("# %q exist\n" , orig)
						fmt.Printf("rm %q\n", path)
					}
					removedCount++
				} else {
					hashedMap[size][hashCur] = path
				}
				return nil
			}
			// 이후에는 hashedMap 사용
			hashCur, err := getFileHash(path, *useCRC32)
			if err != nil {
				log.Printf("Error hashing %s: %v", path, err)
				return nil
			}
			if origPath, exists := hashedMap[size][hashCur]; exists {
				if *doDelete {
					os.Remove(path)
					fmt.Printf("Removed duplicate: %s (original: %s)\n", path, origPath)
				} else {
					fmt.Printf("Would remove duplicate: %s (original: %s)\n", path, origPath)
				}
				removedCount++
			} else {
				hashedMap[size][hashCur] = path
			}
			return nil
		}
	})
	if err != nil {
		log.Fatalf("Scan error: %v", err)
	}

	if *doDelete {
		fmt.Printf("Finished. Removed %d duplicate files.\n", removedCount)
	} else {
		fmt.Printf("Finished dry run. %d duplicates found.\n", removedCount)
	}
}
