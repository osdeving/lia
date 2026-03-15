package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/willams/lia/internal/parser"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt <file.lia>",
	Short: "Format a LIA source file (in-place by default)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		data, err := os.ReadFile(input)
		if err != nil {
			return err
		}

		// Reuse NormalizeLLMOutput to strip markdown fences and normalize syntax
		formatted := parser.NormalizeLLMOutput(string(data))

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Fprint(cmd.OutOrStdout(), formatted)
			return nil
		}

		if err := os.WriteFile(input, []byte(formatted), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "formatted %s\n", input)
		return nil
	},
}

func init() {
	fmtCmd.Flags().Bool("dry-run", false, "print formatted code to stdout instead of modifying file")
	rootCmd.AddCommand(fmtCmd)
}
