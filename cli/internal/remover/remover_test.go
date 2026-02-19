package remover

import (
	"testing"

	"github.com/KashifKhn/nvim-remove-comments/cli/internal/parser"
)

func TestRemove_NoComments(t *testing.T) {
	src := []byte("package main\n\nfunc main() {}\n")
	got := Remove(src, nil)
	if string(got) != string(src) {
		t.Errorf("got %q, want %q", got, src)
	}
}

func TestRemove_FullLineComment(t *testing.T) {
	src := []byte("package main\n// comment\nfunc main() {}\n")
	ranges := []parser.CommentRange{
		{StartRow: 1, StartCol: 0, EndRow: 1, EndCol: 10, IsFullLine: true},
	}
	want := "package main\nfunc main() {}\n"
	got := Remove(src, ranges)
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRemove_InlineComment(t *testing.T) {
	src := []byte("x := 1 // comment\n")
	ranges := []parser.CommentRange{
		{StartRow: 0, StartCol: 7, EndRow: 0, EndCol: 18},
	}
	want := "x := 1\n"
	got := Remove(src, ranges)
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRemove_MultiLineBlock(t *testing.T) {
	src := []byte("a := 1\n/*\nblock\ncomment\n*/\nb := 2\n")
	ranges := []parser.CommentRange{
		{StartRow: 1, StartCol: 0, EndRow: 4, EndCol: 2, IsMultiLine: true},
	}
	want := "a := 1\nb := 2\n"
	got := Remove(src, ranges)
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRemove_MultipleFullLineComments(t *testing.T) {
	src := []byte("// one\n// two\n// three\nfunc f() {}\n")
	ranges := []parser.CommentRange{
		{StartRow: 0, StartCol: 0, EndRow: 0, EndCol: 6, IsFullLine: true},
		{StartRow: 1, StartCol: 0, EndRow: 1, EndCol: 6, IsFullLine: true},
		{StartRow: 2, StartCol: 0, EndRow: 2, EndCol: 8, IsFullLine: true},
	}
	want := "func f() {}\n"
	got := Remove(src, ranges)
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRemove_EmptySource(t *testing.T) {
	got := Remove([]byte{}, nil)
	if len(got) != 0 {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestRemove_InlineComment_TrailingSpaceTrimmed(t *testing.T) {
	src := []byte("x := 1    // comment\n")
	ranges := []parser.CommentRange{
		{StartRow: 0, StartCol: 10, EndRow: 0, EndCol: 20},
	}
	want := "x := 1\n"
	got := Remove(src, ranges)
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRemove_CommentOnlyLine_BecomesEmpty_Dropped(t *testing.T) {
	src := []byte("    // comment\n")
	ranges := []parser.CommentRange{
		{StartRow: 0, StartCol: 4, EndRow: 0, EndCol: 14},
	}
	got := Remove(src, ranges)
	if len(got) != 0 {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestSplitLines_Basic(t *testing.T) {
	tests := []struct {
		input []byte
		want  int
	}{
		{[]byte("a\nb\n"), 3},
		{[]byte("abc"), 1},
		{[]byte("a\n"), 2},
		{[]byte{}, 0},
	}
	for _, tt := range tests {
		got := splitLines(tt.input)
		if len(got) != tt.want {
			t.Errorf("input %q: got %d lines, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestTrimRight(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello   ", "hello"},
		{"hello\t", "hello"},
		{"hello", "hello"},
		{"  ", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := trimRight([]byte(tt.input))
		if string(got) != tt.want {
			t.Errorf("input %q: got %q, want %q", tt.input, got, tt.want)
		}
	}
}
