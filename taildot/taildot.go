package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func main() {
	lines := make(chan string)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return
			}
			fmt.Println(line)
			ticker.Reset(time.Second * 30)
		case <-ticker.C:
			fmt.Println(".")
		}
	}
}
