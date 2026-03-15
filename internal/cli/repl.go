package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/willams/lia/internal/parser"
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start LIA interactive mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "LIA Interactive REPL")
		fmt.Fprintln(cmd.OutOrStdout(), "Type your LIA declarations. Type 'exit' to quit.")

		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Fprint(cmd.OutOrStdout(), "> ")
			if !scanner.Scan() {
				break
			}
			text := strings.TrimSpace(scanner.Text())
			if text == "exit" || text == "quit" {
				break
			}
			if text == "" {
				continue
			}

			// Parse single line or block
			res := parser.ParseWithRecovery(text)
			if res.HasErrors() {
				for _, d := range res.Diagnostics {
					fmt.Fprintf(cmd.ErrOrStderr(), "error: %s at line %d:%d\n", d.Message, d.Line, d.Column)
				}
			} else if res.Program != nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Ok.")
			}
		}
		return scanner.Err()
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
}
