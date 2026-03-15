package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/ir"
)

func printDiagsWithWriter(w io.Writer, diags []ir.Diagnostic) {
	for _, d := range diags {
		if d.Path != "" {
			fmt.Fprintf(w, "%s: %s (%s)\n", d.Severity, d.Message, d.Path)
			continue
		}
		fmt.Fprintf(w, "%s: %s\n", d.Severity, d.Message)
	}
}

func printDiags(cmd *cobra.Command, diags []ir.Diagnostic) {
	printDiagsWithWriter(cmd.OutOrStdout(), diags)
}
