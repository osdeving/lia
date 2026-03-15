package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/parser"
)

type ParseOptions struct {
	Input  string
	Out    string
	Stdout io.Writer
}

func parseParseOptions(cmd *cobra.Command, args []string) ParseOptions {
	out, _ := cmd.Flags().GetString("out")
	if out == "" {
		out = replaceExt(args[0], ".liao")
	}
	return ParseOptions{
		Input:  args[0],
		Out:    out,
		Stdout: cmd.OutOrStdout(),
	}
}

func runParse(opts ParseOptions) error {
	prog, err := parser.ParseFile(opts.Input)
	if err != nil {
		return err
	}
	if err := codec.WriteProgramFile(opts.Out, prog); err != nil {
		return err
	}
	fmt.Fprintf(opts.Stdout, "wrote %s\n", opts.Out)
	return nil
}

var parseCmd = &cobra.Command{
	Use:   "parse <file.lia>",
	Short: "Parse .lia source into a .liao object file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runParse(parseParseOptions(cmd, args))
	},
}

func init() {
	parseCmd.Flags().StringP("out", "o", "", "output .liao file")
}
