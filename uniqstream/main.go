package main

import (
	"bufio"
	"log"
	"os"
)

var cachemap = make(map[string]struct{})

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if _, ok := cachemap[text]; ok {
			continue
		}
		log.Println(text)
		cachemap[text] = struct{}{}
	}
}
