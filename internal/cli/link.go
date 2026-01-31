package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/check"
	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/linker"
)

var linkCmd = &cobra.Command{
	Use:   "link <inputs...>",
	Short: "Link one or more .liao files into a .lial unit",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out, _ := cmd.Flags().GetString("out")
		if out == "" {
			out = "out.lial"
		}
		logPath, _ := cmd.Flags().GetString("decision-log")
		if logPath == "" {
			logPath = out + ".decision-log.json"
		}

		var inputs []*ir.Program
		for _, path := range args {
			p, err := loadProgram(path)
			if err != nil {
				return err
			}
			inputs = append(inputs, p)
		}

		var allPacks []ir.Pack
		var allDiags []ir.Diagnostic
		for _, p := range inputs {
			packs, packDiags := loadPacksForProgram(cmd, p)
			allDiags = append(allDiags, packDiags...)
			allPacks = append(allPacks, packs...)
		}

		linked, log, diags, err := linker.Link(inputs, allPacks)
		if err != nil {
			return err
		}
		allDiags = append(allDiags, diags...)
		printDiags(cmd, allDiags)
		if check.HasErrors(allDiags) {
			return fmt.Errorf("link failed")
		}
		if err := codec.WriteProgramFile(out, linked); err != nil {
			return err
		}
		if err := linker.WriteDecisionLog(logPath, log); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", out)
		fmt.Fprintf(cmd.OutOrStdout(), "decision log %s\n", logPath)
		return nil
	},
}

func init() {
	linkCmd.Flags().StringP("out", "o", "", "output .lial file")
	linkCmd.Flags().String("decision-log", "", "decision log output path")
	addPackDirFlag(linkCmd)
}
