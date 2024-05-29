package main

import (
	"log"
	"os"

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
		// Run: func(cmd *cobra.Command, args []string) {
		// 	// if versionFlag {
		// 	// 	fmt.Println(version)
		// 	// } else {
		// 	// 	cmd.Help()
		// 	// }
		// },
	}

	f := cmd.PersistentFlags()
	f.BoolP("verbose", "v", false, "verbose output for debugging purposes")
	f.BoolVar(&versionFlag, "version", false, "print version")
	f.Parse(args)

	cmd.AddCommand(
		cmdSizecmp(),
	)

	return cmd, nil
}
