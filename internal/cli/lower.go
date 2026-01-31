package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/lower/java"
	"github.com/willams/lia/internal/lower/python"
)

var lowerCmd = &cobra.Command{
	Use:   "lower <file.lial>",
	Short: "Lower a linked unit to a target language (stub)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		target, _ := cmd.Flags().GetString("target")
		out, _ := cmd.Flags().GetString("out")
		if out == "" {
			switch target {
			case "python":
				out = replaceExt(input, ".py")
			default:
				out = replaceExt(input, ".java")
			}
		}

		prog, err := loadProgram(input)
		if err != nil {
			return err
		}

		var data []byte
		switch target {
		case "java":
			data, err = java.Lower(prog)
		case "python":
			data, err = python.Lower(prog)
		default:
			return fmt.Errorf("unknown target: %s", target)
		}
		if err != nil {
			return err
		}

		if err := os.WriteFile(out, data, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", out)
		return nil
	},
}

func init() {
	lowerCmd.Flags().String("target", "java", "target language: java|python")
	lowerCmd.Flags().StringP("out", "o", "", "output file")
}
