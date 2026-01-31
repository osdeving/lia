package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/repro"
)

var replayCmd = &cobra.Command{
	Use:   "replay <file.lial>",
	Short: "Replay a linked unit using a prompt tape (stub)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tapePath, _ := cmd.Flags().GetString("tape")
		if tapePath == "" {
			return fmt.Errorf("--tape is required")
		}
		if _, err := repro.LoadTape(tapePath); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "replay stub: tape loaded, replay not implemented yet")
		return nil
	},
}

func init() {
	replayCmd.Flags().String("tape", "", "prompt tape JSON file")
}
