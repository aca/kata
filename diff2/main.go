package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"sync"
)

type DiffFile struct {
	Head []byte
	Tail []byte
	Size int64
}

var hex2string = hex.EncodeToString

func main() {

	wg := sync.WaitGroup{}

	var dfile = make([]*DiffFile, 2)

	for i := range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			dfile[i], err = GetDiffFile(os.Args[1+i])
			if err != nil {
				panic(err)
			}
		}()
	}
	wg.Wait()

	fmt.Println(os.Args[1])
	fmt.Println("Head:", hex2string(dfile[0].Head[0:min(len(dfile[0].Head), 16)]))
	fmt.Println("Tail:", hex2string(dfile[0].Tail[0:min(len(dfile[0].Tail), 16)]))
	fmt.Println("")
	fmt.Println(os.Args[2])
	fmt.Println("Head:", hex2string(dfile[1].Head[0:min(len(dfile[1].Head), 16)]))
	fmt.Println("Tail:", hex2string(dfile[1].Tail[0:min(len(dfile[1].Tail), 16)]))

	fmt.Println("")
	fmt.Println("Size:", dfile[0].Size, dfile[1].Size)

	if dfile[0].Size != dfile[1].Size || !reflect.DeepEqual(dfile[0].Head, dfile[1].Head) || !reflect.DeepEqual(dfile[0].Tail, dfile[1].Tail) {
		fmt.Println("Files are different")
		os.Exit(1)
	} else {
		fmt.Println("Files are same")
	}
}

const diffsize = 2048

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

	if fsize < diffsize {
		buf, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		return &DiffFile{
			Head: buf,
			Tail: buf,
			Size: fsize,
		}, nil
	}

	start := make([]byte, diffsize)
	_, err = f.Read(start)
	if err != nil {
		return nil, err
	}

	ret, err := f.Seek(-diffsize, 2)
	if err != nil {
		return nil, err
	}

	log.Println("ret:", ret)

	end := make([]byte, diffsize)
	_, err = f.Read(end)
	if err != nil {
		return nil, err
	}

	return &DiffFile{
		Head: start,
		Tail: end,
		Size: fsize,
	}, nil
}
