package output

import (
	"fmt"
	"io"

	"github.com/fatih/color"

	"github.com/KashifKhn/nvim-remove-comments/cli/internal/diff"
)

var (
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
	bold   = color.New(color.Bold)
)

type Printer struct {
	w     io.Writer
	quiet bool
	write bool
}

func New(w io.Writer, quiet, write bool) *Printer {
	return &Printer{w: w, quiet: quiet, write: write}
}

func (p *Printer) File(r diff.Result) {
	if p.quiet || !r.Changed {
		return
	}
	action := "would remove"
	if p.write {
		action = "removed"
	}
	removed := r.LinesRemoved()
	noun := "line"
	if removed != 1 {
		noun = "lines"
	}
	_, _ = yellow.Fprintf(p.w, "  %s  ", action)
	_, _ = fmt.Fprintf(p.w, "%d comment %s from ", removed, noun)
	_, _ = bold.Fprintf(p.w, "%s\n", r.Path)
}

func (p *Printer) Skipped(path, reason string) {
	if p.quiet {
		return
	}
	_, _ = fmt.Fprintf(p.w, "  skip  %s (%s)\n", path, reason)
}

func (p *Printer) Error(path string, err error) {
	_, _ = red.Fprintf(p.w, "  error  %s: %v\n", path, err)
}

func (p *Printer) Summary(changed, skipped, errors, total int) {
	action := "would be modified"
	if p.write {
		action = "modified"
	}
	_, _ = bold.Fprintf(p.w, "\n%d/%d files %s", changed, total, action)
	if skipped > 0 {
		_, _ = fmt.Fprintf(p.w, ", %d skipped", skipped)
	}
	if errors > 0 {
		_, _ = red.Fprintf(p.w, ", %d errors", errors)
	}
	_, _ = fmt.Fprintln(p.w)
}
