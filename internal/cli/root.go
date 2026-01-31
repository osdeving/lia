package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const Version = "0.1.0-dev"

var rootCmd = &cobra.Command{
	Use:           "lia",
	Short:         "LIA toolchain (model-first IR)",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(replayCmd)
	rootCmd.AddCommand(lowerCmd)
}
