package parser

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/KashifKhn/remove-comments/cli/internal/languages"
)

type CommentRange struct {
	StartRow    uint32
	StartCol    uint32
	EndRow      uint32
	EndCol      uint32
	IsFullLine  bool
	IsMultiLine bool
}

func Parse(src []byte, cfg languages.LangConfig) ([]CommentRange, error) {
	lang := cfg.Language()

	p := sitter.NewParser()
	defer p.Close()
	p.SetLanguage(lang)

	tree, err := p.ParseCtx(context.Background(), nil, src)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	defer tree.Close()

	q, err := sitter.NewQuery([]byte(cfg.Query), lang)
	if err != nil {
		return nil, fmt.Errorf("query compile: %w", err)
	}
	defer q.Close()

	lines := splitLines(src)

	qc := sitter.NewQueryCursor()
	defer qc.Close()
	qc.Exec(q, tree.RootNode())

	var ranges []CommentRange
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, cap := range m.Captures {
			n := cap.Node
			sr := n.StartPoint().Row
			sc := n.StartPoint().Column
			er := n.EndPoint().Row
			ec := n.EndPoint().Column

			isMultiLine := er > sr
			isFullLine := false
			if !isMultiLine {
				if int(sr) < len(lines) {
					lineLen := uint32(len(lines[sr]))
					isFullLine = sc == 0 && ec >= lineLen
				}
			}

			ranges = append(ranges, CommentRange{
				StartRow:    sr,
				StartCol:    sc,
				EndRow:      er,
				EndCol:      ec,
				IsFullLine:  isFullLine,
				IsMultiLine: isMultiLine,
			})
		}
	}

	return ranges, nil
}

func splitLines(src []byte) []string {
	var lines []string
	start := 0
	for i, b := range src {
		if b == '\n' {
			lines = append(lines, string(src[start:i]))
			start = i + 1
		}
	}
	if start <= len(src) {
		lines = append(lines, string(src[start:]))
	}
	return lines
}
