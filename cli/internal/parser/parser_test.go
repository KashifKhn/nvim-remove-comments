package parser

import (
	"testing"

	"github.com/KashifKhn/nvim-remove-comments/cli/internal/languages"
)

func langFor(ext string, t *testing.T) languages.LangConfig {
	t.Helper()
	cfg, ok := languages.Get(ext)
	if !ok {
		t.Fatalf("no language config for extension %s", ext)
	}
	return cfg
}

func TestParse_Go_FullLineComment(t *testing.T) {
	src := []byte("package main\n\n// this is a comment\nfunc main() {}\n")
	ranges, err := Parse(src, langFor(".go", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
	r := ranges[0]
	if r.StartRow != 2 {
		t.Errorf("expected StartRow=2, got %d", r.StartRow)
	}
	if !r.IsFullLine {
		t.Error("expected IsFullLine=true")
	}
	if r.IsMultiLine {
		t.Error("expected IsMultiLine=false")
	}
}

func TestParse_Go_BlockComment(t *testing.T) {
	src := []byte("package main\n\n/*\nblock\ncomment\n*/\nfunc f() {}\n")
	ranges, err := Parse(src, langFor(".go", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
	r := ranges[0]
	if !r.IsMultiLine {
		t.Error("expected IsMultiLine=true for block comment")
	}
	if r.StartRow != 2 {
		t.Errorf("expected StartRow=2, got %d", r.StartRow)
	}
	if r.EndRow != 5 {
		t.Errorf("expected EndRow=5, got %d", r.EndRow)
	}
}

func TestParse_Go_InlineComment(t *testing.T) {
	src := []byte("package main\n\nfunc f() {} // inline comment\n")
	ranges, err := Parse(src, langFor(".go", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
	r := ranges[0]
	if r.IsFullLine {
		t.Error("expected IsFullLine=false for inline comment")
	}
	if r.IsMultiLine {
		t.Error("expected IsMultiLine=false")
	}
}

func TestParse_Go_NoComments(t *testing.T) {
	src := []byte("package main\n\nfunc main() {}\n")
	ranges, err := Parse(src, langFor(".go", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 0 {
		t.Errorf("expected 0 comments, got %d", len(ranges))
	}
}

func TestParse_Go_CommentInString_NotRemoved(t *testing.T) {
	src := []byte(`package main

import "fmt"

func main() {
	s := "// this is not a comment"
	fmt.Println(s)
}
`)
	ranges, err := Parse(src, langFor(".go", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 0 {
		t.Errorf("comment-like text inside string should not be detected, got %d range(s)", len(ranges))
	}
}

func TestParse_Python_HashComment(t *testing.T) {
	src := []byte("x = 1\n# full line comment\ny = 2\n")
	ranges, err := Parse(src, langFor(".py", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
	if ranges[0].StartRow != 1 {
		t.Errorf("expected StartRow=1, got %d", ranges[0].StartRow)
	}
}

func TestParse_Python_InlineComment(t *testing.T) {
	src := []byte("x = 1 # inline\n")
	ranges, err := Parse(src, langFor(".py", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
	if ranges[0].IsFullLine {
		t.Error("expected IsFullLine=false for inline python comment")
	}
}

func TestParse_JavaScript_LineAndBlock(t *testing.T) {
	src := []byte("const x = 1;\n// line comment\n/* block */\nconst y = 2;\n")
	ranges, err := Parse(src, langFor(".js", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(ranges))
	}
}

func TestParse_TypeScript_Comment(t *testing.T) {
	src := []byte("const x: number = 1;\n// ts comment\nconst y = 2;\n")
	ranges, err := Parse(src, langFor(".ts", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
}

func TestParse_TSX_Comment(t *testing.T) {
	src := []byte("// tsx comment\nconst C = () => <div/>;\n")
	ranges, err := Parse(src, langFor(".tsx", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
}

func TestParse_JSX_Comment(t *testing.T) {
	src := []byte("// jsx comment\nconst C = () => <div/>;\n")
	ranges, err := Parse(src, langFor(".jsx", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
}

func TestParse_Java_LineAndBlockComment(t *testing.T) {
	src := []byte("class Main {\n// line\n/* block */\npublic static void main(String[] args) {}\n}\n")
	ranges, err := Parse(src, langFor(".java", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(ranges))
	}
}

func TestParse_C_LineAndBlockComment(t *testing.T) {
	src := []byte("// line\nint x = 1;\n/* block\ncomment */\n")
	ranges, err := Parse(src, langFor(".c", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(ranges))
	}
}

func TestParse_Cpp_Comment(t *testing.T) {
	src := []byte("// line\nint x = 1;\n/* block */\n")
	ranges, err := Parse(src, langFor(".cpp", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(ranges))
	}
}

func TestParse_Rust_LineComment(t *testing.T) {
	src := []byte("// rust comment\nfn main() {}\n")
	ranges, err := Parse(src, langFor(".rs", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
}

func TestParse_Lua_Comment(t *testing.T) {
	src := []byte("local x = 1\n-- lua comment\nlocal y = 2\n")
	ranges, err := Parse(src, langFor(".lua", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
}

func TestParse_HTML_Comment(t *testing.T) {
	src := []byte("<!-- html comment -->\n<div>content</div>\n")
	ranges, err := Parse(src, langFor(".html", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
}

func TestParse_CSS_Comment(t *testing.T) {
	src := []byte("/* css comment */\nbody { color: red; }\n")
	ranges, err := Parse(src, langFor(".css", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
}

func TestParse_YAML_Comment(t *testing.T) {
	src := []byte("key: value\n# yaml comment\nother: val\n")
	ranges, err := Parse(src, langFor(".yaml", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
}

func TestParse_TOML_Comment(t *testing.T) {
	src := []byte("[section]\n# toml comment\nkey = \"val\"\n")
	ranges, err := Parse(src, langFor(".toml", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
}

func TestParse_Bash_Comment(t *testing.T) {
	src := []byte("#!/bin/bash\n# bash comment\necho hello\n")
	ranges, err := Parse(src, langFor(".sh", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) < 1 {
		t.Fatalf("expected at least 1 comment, got %d", len(ranges))
	}
}

func TestParse_EmptyFile(t *testing.T) {
	ranges, err := Parse([]byte{}, langFor(".go", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 0 {
		t.Errorf("expected 0 ranges for empty file, got %d", len(ranges))
	}
}

func TestParse_MultipleComments_CorrectCount(t *testing.T) {
	src := []byte("// one\n// two\n// three\nfunc f() {}\n")
	ranges, err := Parse(src, langFor(".go", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 3 {
		t.Fatalf("expected 3 comments, got %d", len(ranges))
	}
}

func TestParse_Go_InlineSingleLineBlock(t *testing.T) {
	src := []byte("package main\n\nfunc f() { /* inline block */ return }\n")
	ranges, err := Parse(src, langFor(".go", t))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(ranges))
	}
	if ranges[0].IsFullLine {
		t.Error("inline block comment should not be IsFullLine")
	}
	if ranges[0].IsMultiLine {
		t.Error("single-line block comment should not be IsMultiLine")
	}
}

func TestSplitLines_BasicNewlines(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []string
	}{
		{"empty", []byte{}, []string{""}},
		{"single line no newline", []byte("hello"), []string{"hello"}},
		{"single line with newline", []byte("hello\n"), []string{"hello", ""}},
		{"two lines", []byte("a\nb\n"), []string{"a", "b", ""}},
		{"three lines no trailing newline", []byte("a\nb\nc"), []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d lines, want %d: %v", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
