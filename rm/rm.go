package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/xtdlib/go-trash"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:                   "rm",
		UseShortOptionHandling: true,
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
			interactive := cmd.Bool("interactive")
			verbose := cmd.Bool("verbose")

			reader := bufio.NewReader(os.Stdin)

			for _, f := range cmd.Args().Slice() {
				if interactive {
					fmt.Printf("rmx: remove %s? ", f)
					_, err := reader.ReadString('\n')
					if err != nil {
						return err
					}
				}

				if _, err := os.Stat(f); os.IsNotExist(err) {
					slog.Debug(fmt.Sprintf("rmx: %q does not exist, skipping", f))
					return nil
				}

				err := trash.Trash(f)
				if err != nil {
					return err
				}

				if verbose {
					slog.Info(fmt.Sprintf("rmx: trashed %q", f))
				}
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
