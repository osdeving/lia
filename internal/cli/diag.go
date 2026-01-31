package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/ir"
)

func printDiags(cmd *cobra.Command, diags []ir.Diagnostic) {
	for _, d := range diags {
		if d.Path != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s (%s)\n", d.Severity, d.Message, d.Path)
			continue
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", d.Severity, d.Message)
	}
}
