package cmd

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/spf13/cobra"

	"github.com/KashifKhn/nvim-remove-comments/cli/internal/diff"
	"github.com/KashifKhn/nvim-remove-comments/cli/internal/output"
	"github.com/KashifKhn/nvim-remove-comments/cli/internal/parser"
	"github.com/KashifKhn/nvim-remove-comments/cli/internal/remover"
	"github.com/KashifKhn/nvim-remove-comments/cli/internal/walker"
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
	flagDiff        bool
	flagLang        string
	flagJobs        int
	flagMaxFileSize int64
)

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&flagWrite, "write", "w", false, "Write changes to disk (default is dry-run)")
	rootCmd.Flags().BoolVarP(&flagQuiet, "quiet", "q", false, "Print only the final summary line")
	rootCmd.Flags().BoolVarP(&flagDiff, "diff", "d", false, "Show unified diff for each changed file")
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
		return err
	}

	jobs := flagJobs
	if jobs <= 0 {
		jobs = runtime.NumCPU()
	}

	entries, walkErrs := walker.Walk(root, flagLang, flagMaxFileSize)
	if len(walkErrs) > 0 {
		for _, e := range walkErrs {
			fmt.Fprintf(os.Stderr, "walk error: %v\n", e)
		}
	}

	printer := output.New(os.Stdout, flagQuiet, flagWrite, flagDiff)

	var (
		mu      sync.Mutex
		changed int32
		errors  int32
		total   int32
	)

	work := make(chan walker.FileEntry, jobs*2)
	var wg sync.WaitGroup

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range work {
				atomic.AddInt32(&total, 1)

				src, err := os.ReadFile(entry.Path)
				if err != nil {
					atomic.AddInt32(&errors, 1)
					mu.Lock()
					printer.Error(entry.Path, err)
					mu.Unlock()
					continue
				}

				ranges, err := parser.Parse(src, entry.Lang)
				if err != nil {
					atomic.AddInt32(&errors, 1)
					mu.Lock()
					printer.Error(entry.Path, err)
					mu.Unlock()
					continue
				}

				after := remover.Remove(src, ranges)
				result := diff.Compute(entry.Path, src, after)

				if result.Changed {
					atomic.AddInt32(&changed, 1)
					if flagWrite {
						if writeErr := os.WriteFile(entry.Path, result.After, 0o644); writeErr != nil {
							atomic.AddInt32(&errors, 1)
							mu.Lock()
							printer.Error(entry.Path, writeErr)
							mu.Unlock()
							continue
						}
					}
				}

				mu.Lock()
				printer.File(result)
				mu.Unlock()
			}
		}()
	}

	for _, e := range entries {
		work <- e
	}
	close(work)
	wg.Wait()

	printer.Summary(int(changed), 0, int(errors), int(total))
	return nil
}
