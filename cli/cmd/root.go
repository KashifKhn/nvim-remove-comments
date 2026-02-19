package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "remove-comments [path]",
	Short: "Remove all comments from source files in a directory",
	Args:  cobra.MaximumNArgs(1),
	RunE:  run,
}

var (
	flagWrite       bool
	flagQuiet       bool
	flagLang        string
	flagJobs        int
	flagMaxFileSize int64
)

func init() {
	rootCmd.Flags().BoolVarP(&flagWrite, "write", "w", false, "Write changes to disk (default is dry-run)")
	rootCmd.Flags().BoolVarP(&flagQuiet, "quiet", "q", false, "Print only the final summary line")
	rootCmd.Flags().StringVar(&flagLang, "lang", "", "Only process files of this language (e.g. go, python)")
	rootCmd.Flags().IntVarP(&flagJobs, "jobs", "j", 0, "Number of parallel workers (default: NumCPU)")
	rootCmd.Flags().Int64Var(&flagMaxFileSize, "max-file-size", 10*1024*1024, "Skip files larger than this size in bytes")
}

func run(cmd *cobra.Command, args []string) error {
	root := "."
	if len(args) == 1 {
		root = args[0]
	}

	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("path %q not found: %w", root, err)
	}

	fmt.Fprintf(os.Stderr, "remove-comments: not yet implemented (root=%s)\n", root)
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
