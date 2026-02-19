package diff

import (
	"testing"
)

func TestCompute_Unchanged(t *testing.T) {
	r := Compute("foo.go", []byte("a\nb\n"), []byte("a\nb\n"))
	if r.Changed {
		t.Error("expected Changed=false for identical content")
	}
	if r.LinesRemoved() != 0 {
		t.Errorf("expected 0 lines removed, got %d", r.LinesRemoved())
	}
}

func TestCompute_Changed(t *testing.T) {
	before := []byte("a\n// comment\nb\n")
	after := []byte("a\nb\n")
	r := Compute("foo.go", before, after)
	if !r.Changed {
		t.Error("expected Changed=true")
	}
	if r.LinesRemoved() != 1 {
		t.Errorf("expected 1 line removed, got %d", r.LinesRemoved())
	}
}

func TestCompute_Unified_Empty_WhenUnchanged(t *testing.T) {
	r := Compute("foo.go", []byte("x\n"), []byte("x\n"))
	if r.Unified() != "" {
		t.Errorf("expected empty unified diff for unchanged, got %q", r.Unified())
	}
}

func TestCompute_Unified_NonEmpty_WhenChanged(t *testing.T) {
	before := []byte("// comment\nx := 1\n")
	after := []byte("x := 1\n")
	r := Compute("foo.go", before, after)
	u := r.Unified()
	if u == "" {
		t.Error("expected non-empty unified diff")
	}
	if len(u) < 10 {
		t.Errorf("unified diff too short: %q", u)
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		input []byte
		want  int
	}{
		{[]byte{}, 0},
		{[]byte("a\n"), 1},
		{[]byte("a\nb\n"), 2},
		{[]byte("a\nb"), 2},
		{[]byte("a"), 1},
	}
	for _, tt := range tests {
		got := countLines(tt.input)
		if got != tt.want {
			t.Errorf("input %q: got %d, want %d", tt.input, got, tt.want)
		}
	}
}
