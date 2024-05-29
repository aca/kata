package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type funcSizecmp struct {
	FileA string
	FileB string
	X     string
}

func (c *funcSizecmp) Run() error {
	sz1, err := readSize(c.FileA)
	if err != nil {
		return err
	}
	sz2, err := readSize(c.FileB)
	if err != nil {
		return err
	}

	if sz1 == sz2 {
		return nil
	}
	return fmt.Errorf("size mismatch: %s: %d, %s: %d", c.FileA, sz1, c.FileB, sz2)
}

func cmdSizecmp() *cobra.Command {
	f := &funcSizecmp{}

	cmd := &cobra.Command{
		Use: "sizecmp",
		// SilenceUsage:  true,
		// SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.FileA = args[0]
			f.FileB = args[1]
			return f.Run()
		},
		// ValidArgsFunction: SecretCompletion(ctx),
	}

	flags := cmd.Flags()
	_ = flags
	flags.StringVarP(&f.X, "xxx", "x", "default", "Description")

	return cmd
}

func readSize(file string) (int64, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}
