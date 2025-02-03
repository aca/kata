package main

import (
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"

	"github.com/spf13/pflag"
)

type DiffFile struct {
	Head []byte
	Tail []byte
	Size int64
	Sum  []byte
}

var hex2string = hex.EncodeToString
var rml bool
var rmr bool
var crc bool

func main() {
	pflag.BoolVar(&rml, "rml", false, "rm left")
	pflag.BoolVar(&rmr, "rmr", false, "rm right")
	pflag.BoolVar(&crc, "crc", false, "use crc32")
	pflag.Parse()
	args := pflag.Args()

	// TODO:should do realpath with evalsymlink
	abs1, err := filepath.Abs(args[0])
	if err != nil {
		panic(err)
	}
	abs2, err := filepath.Abs(args[1])
	if err != nil {
		panic(err)
	}
	if abs1 == abs2 {
		panic("same file")
	}

	wg := sync.WaitGroup{}

	var dfile = make([]*DiffFile, 2)

	for i := range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			dfile[i], err = GetDiffFile(args[i])
			if err != nil {
				panic(err)
			}
		}()
	}
	wg.Wait()

	fmt.Println("file:", args[0])
	fmt.Println("file:", args[1])
	fmt.Printf("Head: %v\n", hex2string(dfile[0].Sum))
	fmt.Printf("Head: %v\n", hex2string(dfile[1].Sum))
	fmt.Println("Size:", dfile[0].Size)
	fmt.Println("Size:", dfile[1].Size)

	defer func() {
		fmt.Println("---")
	}()

	if dfile[0].Size != dfile[1].Size || !reflect.DeepEqual(dfile[0].Sum, dfile[1].Sum) {
		if dfile[0].Size < dfile[1].Size {
			fmt.Printf("%q is larger\n", args[1])
		}
		if dfile[0].Size > dfile[1].Size {
			fmt.Printf("%q is larger\n", args[0])
		}
		fmt.Println("different hash")
	} else {
		fmt.Printf("%v = %v\n", args[0], args[1])
		if rml {
			fmt.Println("rm", args[0])
			err := os.Remove(args[0])
			if err != nil {
				log.Fatal(err)
			}
		}
		if rmr {
			fmt.Println("rm", args[1])
			err := os.Remove(args[1])
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

const diffsize = 2048

// GetFileCRC32 calculates and returns the CRC32 checksum of the specified file.
func GetFileCRC32(f *os.File) (uint32) {
	defer f.Close()

	// Create a new CRC32 hasher (IEEE polynomial)
	hasher := crc32.NewIEEE()

	// Copy the file's data into the hasher
	if _, err := io.Copy(hasher, f); err != nil {
		panic(err)
	}

	// Compute the checksum
	sum := hasher.Sum32()
	return sum
}

func GetDiffFile(fname string) (*DiffFile, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	fsize := fi.Size()

	if crc {
		return &DiffFile{
			Size: fsize,
			Sum: []byte(fmt.Sprintf("%x", GetFileCRC32(f))),
		}, nil
	}

	if fsize < diffsize {
		buf, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		return &DiffFile{
			Sum:  buf,
			Size: fsize,
		}, nil
	}

	{
		start := make([]byte, diffsize)
		_, err = f.Read(start)
		if err != nil {
			return nil, err
		}

		_, err = f.Seek(-diffsize, 2)
		if err != nil {
			return nil, err
		}

		end := make([]byte, diffsize)
		_, err = f.Read(end)
		if err != nil {
			return nil, err
		}

		return &DiffFile{
			Sum:  slices.Concat(start, end),
			Size: fsize,
		}, nil
	}

}
