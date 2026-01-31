package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/parser"
)

var parseCmd = &cobra.Command{
	Use:   "parse <file.lia>",
	Short: "Parse .lia source into a .liao object file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		out, _ := cmd.Flags().GetString("out")
		if out == "" {
			out = replaceExt(input, ".liao")
		}

		prog, err := parser.ParseFile(input)
		if err != nil {
			return err
		}
		if err := codec.WriteProgramFile(out, prog); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", out)
		return nil
	},
}

func init() {
	parseCmd.Flags().StringP("out", "o", "", "output .liao file")
}
