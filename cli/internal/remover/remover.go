package remover

import (
	"github.com/KashifKhn/remove-comments/cli/internal/parser"
)

type inlineEdit struct {
	row      uint32
	startCol uint32
	endCol   uint32
}

func Remove(src []byte, ranges []parser.CommentRange) []byte {
	if len(ranges) == 0 {
		return src
	}

	lines := splitLines(src)
	out := processLines(lines, ranges)
	return joinLines(out)
}

func splitLines(src []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range src {
		if b == '\n' {
			lines = append(lines, src[start:i+1])
			start = i + 1
		}
	}
	if start < len(src) {
		lines = append(lines, src[start:])
	} else if start == len(src) && len(src) > 0 {
		lines = append(lines, []byte{})
	}
	return lines
}

func joinLines(lines [][]byte) []byte {
	var total int
	for _, l := range lines {
		total += len(l)
	}
	out := make([]byte, 0, total)
	for _, l := range lines {
		out = append(out, l...)
	}
	return out
}

func processLines(lines [][]byte, ranges []parser.CommentRange) [][]byte {
	delRows := map[uint32]bool{}
	var inlineEdits []inlineEdit

	for _, r := range ranges {
		if r.IsMultiLine {
			for row := r.StartRow; row <= r.EndRow; row++ {
				delRows[row] = true
			}
			continue
		}
		if r.IsFullLine {
			delRows[r.StartRow] = true
			continue
		}
		inlineEdits = append(inlineEdits, inlineEdit{
			row:      r.StartRow,
			startCol: r.StartCol,
			endCol:   r.EndCol,
		})
	}

	out := make([][]byte, 0, len(lines))
	for i, line := range lines {
		row := uint32(i)
		if delRows[row] {
			continue
		}
		trimmed := applyInlineEdits(line, inlineEdits, row)
		if trimmed == nil {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func applyInlineEdits(line []byte, edits []inlineEdit, row uint32) []byte {
	type span struct {
		start, end uint32
	}
	var spans []span
	for _, e := range edits {
		if e.row == row {
			spans = append(spans, span{e.startCol, e.endCol})
		}
	}
	if len(spans) == 0 {
		return line
	}

	hasNewline := len(line) > 0 && line[len(line)-1] == '\n'
	content := line
	if hasNewline {
		content = line[:len(line)-1]
	}

	result := make([]byte, 0, len(content))
	for i := uint32(0); i < uint32(len(content)); i++ {
		inSpan := false
		for _, s := range spans {
			if i >= s.start && i < s.end {
				inSpan = true
				break
			}
		}
		if !inSpan {
			result = append(result, content[i])
		}
	}

	result = trimRight(result)
	if len(result) == 0 {
		return nil
	}
	if hasNewline {
		result = append(result, '\n')
	}
	return result
}

func trimRight(b []byte) []byte {
	end := len(b)
	for end > 0 && (b[end-1] == ' ' || b[end-1] == '\t') {
		end--
	}
	return b[:end]
}
