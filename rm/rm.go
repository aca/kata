package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/go-stdx/trash"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name: "rm",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Value:   false,
				Aliases: []string{"f"},
			},
			&cli.BoolFlag{
				Name:    "interactive",
				Value:   false,
				Aliases: []string{"i", "I"},
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Value:   false,
				Aliases: []string{"r", "R"},
			},
			&cli.BoolFlag{
				Name:    "dir",
				Value:   false,
				Aliases: []string{"d"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Value:   false,
				Aliases: []string{"v"},
			},
			&cli.BoolFlag{
				Name:  "one-file-system",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "no-preserve-root",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "preserve-root",
				Value: "",
			},
			&cli.BoolFlag{
				Name:  "version",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "help",
				Value: false,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			for _, f := range cmd.Args().Slice() {
				err := trash.Trash(f)
				slog.Info(fmt.Sprintf("trash: trashed %q", f))
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
