package main

import (
	"log"
	"os"

	"github.com/aca/lazybox/sizecmp"
	"github.com/spf13/cobra"
)

func main() {
	rootcmd, err := newRootCmd(os.Args)
	if err != nil {
		panic(err)
	}
	if err := rootcmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func newRootCmd(args []string) (*cobra.Command, error) {
	versionFlag := false
	cmd := &cobra.Command{
		Use:           "lazybox",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	f := cmd.PersistentFlags()
	f.BoolP("verbose", "v", false, "verbose output for debugging purposes")
	f.BoolVar(&versionFlag, "version", false, "print version")
	f.Parse(args)

	cmd.AddCommand(
		sizecmp.Command(),
	)

	return cmd, nil
}
