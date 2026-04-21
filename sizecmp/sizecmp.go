package sizecmp

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type CommandOpt struct {
	FileA string
	FileB string
	X     string
}

func Run(opt *CommandOpt) error {
	sz1, err := readSize(opt.FileA)
	if err != nil {
		return err
	}
	sz2, err := readSize(opt.FileB)
	if err != nil {
		return err
	}

	if sz1 == sz2 {
		return nil
	}
	return fmt.Errorf("size mismatch: %s: %d, %s: %d", opt.FileA, sz1, opt.FileB, sz2)
}

func Command() *cobra.Command {
	f := &CommandOpt{}

	cmd := &cobra.Command{
		Use:  "sizecmp",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.FileA = args[0]
			f.FileB = args[1]
			return Run(f)
		},
	}

	flags := cmd.Flags()
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
