package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/check"
	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/ir"
	"github.com/willams/lia/internal/linker"
)

type LinkOptions struct {
	Inputs   []string
	Out      string
	LogPath  string
	PackDirs []string
	Stdout   io.Writer
	Stderr   io.Writer
}

func parseLinkOptions(cmd *cobra.Command, args []string) LinkOptions {
	out, _ := cmd.Flags().GetString("out")
	if out == "" {
		out = "out.lial"
	}
	logPath, _ := cmd.Flags().GetString("decision-log")
	if logPath == "" {
		logPath = out + ".decision-log.json"
	}

	return LinkOptions{
		Inputs:   args,
		Out:      out,
		LogPath:  logPath,
		PackDirs: resolvePackDirs(cmd),
		Stdout:   cmd.OutOrStdout(),
		Stderr:   cmd.ErrOrStderr(),
	}
}

func runLink(opts LinkOptions) error {
	var inputs []*ir.Program
	for _, path := range opts.Inputs {
		p, err := loadProgram(path)
		if err != nil {
			return err
		}
		inputs = append(inputs, p)
	}

	var allPacks []ir.Pack
	var allDiags []ir.Diagnostic
	for _, p := range inputs {
		packs, packDiags := loadPacksForProgramWithDirs(opts.PackDirs, p)
		allDiags = append(allDiags, packDiags...)
		allPacks = append(allPacks, packs...)
	}

	linked, log, diags, err := linker.Link(inputs, allPacks)
	if err != nil {
		return err
	}
	allDiags = append(allDiags, diags...)
	printDiagsWithWriter(opts.Stderr, allDiags)
	
	if check.HasErrors(allDiags) {
		return fmt.Errorf("link failed")
	}
	if err := codec.WriteProgramFile(opts.Out, linked); err != nil {
		return err
	}
	if err := linker.WriteDecisionLog(opts.LogPath, log); err != nil {
		return err
	}

	fmt.Fprintf(opts.Stdout, "wrote %s\n", opts.Out)
	fmt.Fprintf(opts.Stdout, "decision log %s\n", opts.LogPath)
	return nil
}

var linkCmd = &cobra.Command{
	Use:   "link <inputs...>",
	Short: "Link one or more .liao files into a .lial unit",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLink(parseLinkOptions(cmd, args))
	},
}

func init() {
	linkCmd.Flags().StringP("out", "o", "", "output .lial file")
	linkCmd.Flags().String("decision-log", "", "decision log output path")
	addPackDirFlag(linkCmd)
}
