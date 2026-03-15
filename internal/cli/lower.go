package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/lower/java"
	"github.com/willams/lia/internal/lower/python"
)

type LowerOptions struct {
	Input       string
	Target      string
	ProfileName string
	Strict      bool
	Out         string
	OutDir      string
	Stdout      *os.File
}

func parseLowerOptions(cmd *cobra.Command, args []string) (LowerOptions, error) {
	target, _ := cmd.Flags().GetString("target")
	profileName, _ := cmd.Flags().GetString("profile")
	strict, _ := cmd.Flags().GetBool("strict")
	out, _ := cmd.Flags().GetString("out")
	outDir, _ := cmd.Flags().GetString("out-dir")

	return LowerOptions{
		Input:       args[0],
		Target:      target,
		ProfileName: profileName,
		Strict:      strict,
		Out:         out,
		OutDir:      outDir,
		Stdout:      os.Stdout,
	}, nil
}

func runLower(opts LowerOptions) error {
	prog, err := loadProgram(opts.Input)
	if err != nil {
		return err
	}

	switch opts.Target {
	case "java":
		profile, err := java.ParseProfile(opts.ProfileName)
		if err != nil {
			return err
		}
		if strings.TrimSpace(opts.OutDir) == "" {
			if strings.TrimSpace(opts.Out) != "" {
				opts.OutDir = opts.Out
			} else {
				opts.OutDir = defaultJavaLowerDir(opts.Input)
			}
		}
		project, err := java.LowerProjectWithOptions(prog, java.Options{Profile: profile, Strict: opts.Strict})
		if err != nil {
			return err
		}
		if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
			return err
		}
		if err := java.WriteProject(opts.OutDir, project); err != nil {
			return err
		}
		fmt.Fprintf(opts.Stdout, "wrote %s (%s)\n", opts.OutDir, profile)
		return nil
	case "python":
		if strings.TrimSpace(opts.Out) == "" {
			if strings.TrimSpace(opts.OutDir) != "" {
				opts.Out = opts.OutDir
			} else {
				opts.Out = replaceExt(opts.Input, ".py")
			}
		}
		data, err := python.Lower(prog)
		if err != nil {
			return err
		}
		if err := os.WriteFile(opts.Out, data, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(opts.Stdout, "wrote %s\n", opts.Out)
		return nil
	default:
		return fmt.Errorf("unknown target: %s", opts.Target)
	}
}

var lowerCmd = &cobra.Command{
	Use:   "lower <file.lial>",
	Short: "Lower a linked unit to a target language",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := parseLowerOptions(cmd, args)
		if err != nil {
			return err
		}
		return runLower(opts)
	},
}

func init() {
	lowerCmd.Flags().String("target", "java", "target language: java|python")
	lowerCmd.Flags().String("profile", "plain", "java lowering profile: plain|spring-boot|quarkus")
	lowerCmd.Flags().Bool("strict", false, "fail instead of generating placeholders or adapter stubs in java lowering")
	lowerCmd.Flags().StringP("out", "o", "", "output file (python) or output directory (java)")
	lowerCmd.Flags().String("out-dir", "", "output directory for multi-file targets")
}

func defaultJavaLowerDir(input string) string {
	base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	if base == "" {
		base = "generated"
	}
	return filepath.Join(filepath.Dir(input), base+"-java")
}
