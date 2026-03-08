package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/check"
	"github.com/willams/lia/internal/repro"
)

var replayCmd = &cobra.Command{
	Use:   "replay <file.lial>",
	Short: "Validate a linked unit against its prompt tape",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tapePath, _ := cmd.Flags().GetString("tape")
		if tapePath == "" {
			return fmt.Errorf("--tape is required")
		}
		prog, err := loadProgram(args[0])
		if err != nil {
			return err
		}
		tape, err := repro.LoadTape(tapePath)
		if err != nil {
			return err
		}
		diags, summary := repro.ValidateProgramAgainstTape(prog, tape)
		printDiags(cmd, diags)
		if check.HasErrors(diags) {
			return fmt.Errorf("replay validation failed")
		}

		fmt.Fprintf(cmd.OutOrStdout(), "tape entries %d\n", summary.TapeEntries)
		fmt.Fprintf(cmd.OutOrStdout(), "modules %d, generated %d, validated %d\n", summary.Modules, summary.GeneratedModules, summary.ValidatedModules)
		fmt.Fprintln(cmd.OutOrStdout(), "ok")
		return nil
	},
}

func init() {
	replayCmd.Flags().String("tape", "", "prompt tape JSON file")
}
