package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Last11CharsWithoutExt returns the last 11 characters of a filename, excluding
// its extension. If the filename (minus extension) is shorter than 11 characters,
// the function returns it as is.
func Last11CharsWithoutExt(filename string) string {
    // Extract the extension (e.g., ".txt", ".go")
    ext := filepath.Ext(filename)
    // Remove the extension from the filename
    base := filename[:len(filename)-len(ext)]

    // If the base name is shorter or equal to 11 characters, return it directly
    if len(base) <= 11 {
        return base
    }

    // Otherwise, return the last 11 characters
    return base[len(base)-11:]
}

func main() {
	fname := os.Args[1]
	fmt.Println(Last11CharsWithoutExt(fname))
}
