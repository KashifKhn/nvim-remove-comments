package diff

import (
	"bytes"
	"fmt"
)

type Result struct {
	Path    string
	Before  []byte
	After   []byte
	Changed bool
}

func Compute(path string, before, after []byte) Result {
	return Result{
		Path:    path,
		Before:  before,
		After:   after,
		Changed: !bytes.Equal(before, after),
	}
}

func (r Result) LinesRemoved() int {
	return countLines(r.Before) - countLines(r.After)
}

func (r Result) Unified() string {
	if !r.Changed {
		return ""
	}
	before := splitLines(r.Before)
	after := splitLines(r.After)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "--- %s\n+++ %s\n", r.Path, r.Path)

	bi, ai := 0, 0
	for bi < len(before) || ai < len(after) {
		if bi < len(before) && ai < len(after) && before[bi] == after[ai] {
			bi++
			ai++
			continue
		}
		if bi < len(before) {
			fmt.Fprintf(&buf, "-%s\n", before[bi])
			bi++
		}
		if ai < len(after) {
			fmt.Fprintf(&buf, "+%s\n", after[ai])
			ai++
		}
	}
	return buf.String()
}

func countLines(b []byte) int {
	if len(b) == 0 {
		return 0
	}
	n := bytes.Count(b, []byte("\n"))
	if b[len(b)-1] != '\n' {
		n++
	}
	return n
}

func splitLines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	raw := bytes.Split(b, []byte("\n"))
	result := make([]string, 0, len(raw))
	for _, l := range raw {
		result = append(result, string(l))
	}
	if len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	return result
}
