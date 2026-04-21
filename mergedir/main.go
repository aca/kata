package main

import (
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
)

// mergeDirectories moves the contents of srcDir into destDir, handling conflicts with size and CRC32 checks
func mergeDirectories(srcDir, destDir string) error {
	const maxSizeForHash = 100 * 1024 * 1024 * 1000 // 100MB in bytes

	// Walk through the source directory
	err := filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if srcPath == srcDir {
			return nil
		}

		// Calculate the relative path and corresponding destination path
		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %v", err)
		}
		destPath := filepath.Join(destDir, relPath)

		// Handle directories
		if info.IsDir() {
			// Create the directory in the destination if it doesn’t exist
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				err := os.MkdirAll(destPath, info.Mode())
				if err != nil {
					return fmt.Errorf("failed to create directory %s: %v", destPath, err)
				}
				log.Printf("Created directory: %s", destPath)
			}
			return nil
		}

		// Handle files
		// Check if the file already exists in the destination
		destInfo, err := os.Stat(destPath)
		if err == nil {
			// Conflict: file exists
			if info.Size() == destInfo.Size() && info.Size() < maxSizeForHash {
				// Files have the same size and are under 100MB, compare CRC32 hashes
				srcHash, err := computeCRC32(srcPath)
				if err != nil {
					return fmt.Errorf("failed to compute CRC32 for %s: %v", srcPath, err)
				}
				destHash, err := computeCRC32(destPath)
				if err != nil {
					return fmt.Errorf("failed to compute CRC32 for %s: %v", destPath, err)
				}

				if srcHash == destHash {
					// Files are identical, print removal command instead of removing
					log.Printf("Identical files detected, run this to remove source: ```rm %s``` (matches %s)", srcPath, destPath)
					return nil
				}
			}
			// Files differ (size, hash, or too large), log conflict and skip
			log.Printf("Conflict: %s differs from %s skipping", destPath, srcPath)
			log.Printf("src %v %v %v", srcPath, info.Size(), humanize.Bytes(uint64(info.Size())))
			log.Printf("dst %v %v %v", destPath, destInfo.Size(), humanize.Bytes(uint64(destInfo.Size())))
			return nil
		} else if !os.IsNotExist(err) {
			// Some other error occurred
			return fmt.Errorf("failed to check destination path %s: %v", destPath, err)
		}

		// Ensure the parent directory exists in the destination
		destDirPath := filepath.Dir(destPath)
		if _, err := os.Stat(destDirPath); os.IsNotExist(err) {
			err := os.MkdirAll(destDirPath, 0755)
			if err != nil {
				return fmt.Errorf("failed to create parent directory %s: %v", destDirPath, err)
			}
		}

		// Move the file
		err = os.Rename(srcPath, destPath)
		if err != nil {
			return fmt.Errorf("failed to move %s to %s: %v", srcPath, destPath, err)
		}
		log.Printf("Moved: %s -> %s", srcPath, destPath)
		return nil
	})

	return err
}

// computeCRC32 calculates the CRC32 hash of a file
func computeCRC32(filePath string) (uint32, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	hash := crc32.NewIEEE()
	_, err = io.Copy(hash, file)
	if err != nil {
		return 0, err
	}

	return hash.Sum32(), nil
}

func main() {
	// Check command-line arguments
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <source_dir> <dest_dir>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s /xxx/a /xxx/b\n", os.Args[0])
		os.Exit(1)
	}

	srcDir := os.Args[1]
	destDir := os.Args[2]

	// Ensure source directory exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		log.Fatalf("Source directory %s does not exist", srcDir)
	}

	// Ensure destination directory exists, create it if it doesn’t
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err := os.MkdirAll(destDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create destination directory %s: %v", destDir, err)
		}
		log.Printf("Created destination directory: %s", destDir)
	}

	// Perform the merge
	err := mergeDirectories(srcDir, destDir)
	if err != nil {
		log.Fatalf("Error merging directories: %v", err)
	}

	log.Println("Directory merge completed successfully")
}
