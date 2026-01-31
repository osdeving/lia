package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/willams/lia/internal/codec"
	"github.com/willams/lia/internal/linker"
)

var explainCmd = &cobra.Command{
	Use:   "explain <file.liao|file.lial>",
	Short: "Explain a LIA artifact (hashes and counts)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		prog, err := loadProgram(path)
		if err != nil {
			return err
		}
		hash, err := codec.HashProgram(prog)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "hash %s\n", hash)
		fmt.Fprintf(cmd.OutOrStdout(), "projects %d, packs %d, modules %d\n", len(prog.Projects), len(prog.Packs), len(prog.Modules))

		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			for _, m := range prog.Modules {
				fmt.Fprintf(cmd.OutOrStdout(), "module %s\n", m.Name)
			}
		}

		logPath, _ := cmd.Flags().GetString("decision-log")
		if logPath != "" {
			log, err := loadDecisionLog(logPath)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "decision log entries %d\n", len(log.Entries))
			if log.Hash != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "decision log hash %s\n", log.Hash)
			}
		}

		return nil
	},
}

func init() {
	explainCmd.Flags().BoolP("verbose", "v", false, "print module names")
	explainCmd.Flags().String("decision-log", "", "decision log file to summarize")
}

func loadDecisionLog(path string) (*linker.DecisionLog, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	var log linker.DecisionLog
	if err := dec.Decode(&log); err != nil {
		return nil, err
	}
	return &log, nil
}
