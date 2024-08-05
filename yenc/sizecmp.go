package sizecmp

import (
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yenc.v0"
)

type CommandOpt struct {
	Decode bool
	File string
}

func Run(opt *CommandOpt) error {
	if opt.File == "" {
		opt.File = "-"
	}

	return nil
}

func Command() *cobra.Command {
	f := &CommandOpt{}

	cmd := &cobra.Command{
		Use:  "yenc",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(f)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&f.Decode, "decode", "d", false, "Decode yEnc file")

	return cmd
}
