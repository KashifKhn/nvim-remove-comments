package output

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"

	"github.com/KashifKhn/nvim-remove-comments/cli/internal/diff"
)

func init() {
	color.NoColor = true
}

func TestPrinter_File_Quiet_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf, true, false)
	r := diff.Compute("foo.go", []byte("// c\nx\n"), []byte("x\n"))
	p.File(r)
	if buf.Len() != 0 {
		t.Errorf("expected no output in quiet mode, got %q", buf.String())
	}
}

func TestPrinter_File_Unchanged_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf, false, false)
	r := diff.Compute("foo.go", []byte("x\n"), []byte("x\n"))
	p.File(r)
	if buf.Len() != 0 {
		t.Errorf("expected no output for unchanged file, got %q", buf.String())
	}
}

func TestPrinter_File_DryRun_ContainsWouldRemove(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf, false, false)
	r := diff.Compute("foo.go", []byte("// comment\nx\n"), []byte("x\n"))
	p.File(r)
	if !strings.Contains(buf.String(), "would remove") {
		t.Errorf("expected 'would remove' in output, got %q", buf.String())
	}
}

func TestPrinter_File_WriteMode_ContainsRemoved(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf, false, true)
	r := diff.Compute("foo.go", []byte("// comment\nx\n"), []byte("x\n"))
	p.File(r)
	if !strings.Contains(buf.String(), "removed") {
		t.Errorf("expected 'removed' in output, got %q", buf.String())
	}
}

func TestPrinter_Error_AlwaysPrints(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf, true, false)
	p.Error("bar.go", fmt.Errorf("some error"))
	if !strings.Contains(buf.String(), "bar.go") {
		t.Errorf("expected path in error output, got %q", buf.String())
	}
}

func TestPrinter_Summary_ContainsCount(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf, false, false)
	p.Summary(3, 1, 0, 10)
	out := buf.String()
	if !strings.Contains(out, "3/10") {
		t.Errorf("expected '3/10' in summary, got %q", out)
	}
}
