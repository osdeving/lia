package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/check"
)

var checkCmd = &cobra.Command{
	Use:   "check <file.liao>",
	Short: "Validate a .liao or .lial file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		prog, err := loadProgram(input)
		if err != nil {
			return err
		}

		packs, packDiags := loadPacksForProgram(cmd, prog)
		diags := append(packDiags, check.CheckProgram(prog, packs)...)
		printDiags(cmd, diags)

		if check.HasErrors(diags) {
			return fmt.Errorf("validation failed")
		}
		fmt.Fprintln(cmd.OutOrStdout(), "ok")
		return nil
	},
}

func init() {
	addPackDirFlag(checkCmd)
}
