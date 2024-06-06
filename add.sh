#!/usr/bin/env bash
set -euxo pipefail

cmd="$1"
mkdir -p "$cmd"
cat >> "$cmd/$cmd.go" << EOF
package $cmd

import "github.com/spf13/cobra"

type CommandOpt struct {
	FileA string
	FileB string
	X     string
}

func Run(opt *CommandOpt) error {
    return nil
}

func Command() *cobra.Command {
	f := &CommandOpt{}

	cmd := &cobra.Command{
		Use:  "$cmd",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(f)
		},
	}

	// flags := cmd.Flags()
    // flags.StringVarP(&f.X, "xxx", "x", "default", "Description")

	return cmd
}
EOF


