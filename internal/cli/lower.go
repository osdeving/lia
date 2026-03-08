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

var lowerCmd = &cobra.Command{
	Use:   "lower <file.lial>",
	Short: "Lower a linked unit to a target language",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		target, _ := cmd.Flags().GetString("target")
		profileName, _ := cmd.Flags().GetString("profile")
		out, _ := cmd.Flags().GetString("out")
		outDir, _ := cmd.Flags().GetString("out-dir")

		prog, err := loadProgram(input)
		if err != nil {
			return err
		}

		switch target {
		case "java":
			profile, err := java.ParseProfile(profileName)
			if err != nil {
				return err
			}
			if strings.TrimSpace(outDir) == "" {
				if strings.TrimSpace(out) != "" {
					outDir = out
				} else {
					outDir = defaultJavaLowerDir(input)
				}
			}
			project, err := java.LowerProjectWithOptions(prog, java.Options{Profile: profile})
			if err != nil {
				return err
			}
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return err
			}
			if err := java.WriteProject(outDir, project); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s (%s)\n", outDir, profile)
			return nil
		case "python":
			if strings.TrimSpace(out) == "" {
				if strings.TrimSpace(outDir) != "" {
					out = outDir
				} else {
					out = replaceExt(input, ".py")
				}
			}
			data, err := python.Lower(prog)
			if err != nil {
				return err
			}
			if err := os.WriteFile(out, data, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", out)
			return nil
		default:
			return fmt.Errorf("unknown target: %s", target)
		}
	},
}

func init() {
	lowerCmd.Flags().String("target", "java", "target language: java|python")
	lowerCmd.Flags().String("profile", "plain", "java lowering profile: plain|spring-boot|quarkus")
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
