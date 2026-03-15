package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/check"
)

type CheckOptions struct {
	Input    string
	PackDirs []string
	Stdout   io.Writer
	Stderr   io.Writer
}

func parseCheckOptions(cmd *cobra.Command, args []string) (CheckOptions, error) {
	return CheckOptions{
		Input:    args[0],
		PackDirs: resolvePackDirs(cmd),
		Stdout:   cmd.OutOrStdout(),
		Stderr:   cmd.ErrOrStderr(),
	}, nil
}

func runCheck(opts CheckOptions) error {
	prog, err := loadProgram(opts.Input)
	if err != nil {
		return err
	}

	packs, packDiags := loadPacksForProgramWithDirs(opts.PackDirs, prog)
	diags := append(packDiags, check.CheckProgram(prog, packs)...)
	printDiagsWithWriter(opts.Stderr, diags)

	if check.HasErrors(diags) {
		return fmt.Errorf("validation failed")
	}
	fmt.Fprintln(opts.Stdout, "ok")
	return nil
}

var checkCmd = &cobra.Command{
	Use:   "check <file.liao>",
	Short: "Validate a .liao or .lial file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := parseCheckOptions(cmd, args)
		if err != nil {
			return err
		}
		return runCheck(opts)
	},
}

func init() {
	addPackDirFlag(checkCmd)
}
