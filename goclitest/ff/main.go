package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/peterbourgon/ff/v4"
)

func main() {
	fs := flag.NewFlagSet("myprogram", flag.ContinueOnError)
	var (
		listenAddr = fs.String("listen", "localhost:8080", "listen address")
		refresh    = fs.Duration("refresh", 15*time.Second, "refresh interval")
		debug      = fs.Bool("debug", false, "log debug information")
		_          = fs.String("config", "", "config file (optional)")
	)

	ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("MY_PROGRAM"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
	)

	fmt.Printf("listen=%s refresh=%s debug=%v\n", *listenAddr, *refresh, *debug)
}
