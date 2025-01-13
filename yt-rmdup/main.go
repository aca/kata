package main

import (
	"fmt"
	"io/fs"
	"log"
	"os/exec"

	"github.com/xtdlib/filepath"
)

func main() {
	log.Println("start")

	var cache = make(map[string]string)

	err := filepath.WalkDir("/mnt/archive-0/Youtube", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.IsDir(d.Type()) {
			return nil
		}

		fileinfo, err := d.Info()
		if err != nil {
			return err
		}

		// skip small
		if fileinfo.Size() < 10000 {
			return nil
		}

		// skip metadata
		if filepath.Ext(path) == ".json" {
			return nil
		}

		stem := filepath.Stem(path)
		if len(stem) < 11 {
			return fmt.Errorf("invalid stem %s", stem)
		}
		ytid := stem[len(stem)-11:]

		if orig, ok := cache[ytid]; ok {
			log.Println(orig)
			log.Println(path)
			exec.Command("mpv", orig, path).Run()
			log.Println("duplicate", ytid, path)
		} else {
			cache[ytid] = path
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}
