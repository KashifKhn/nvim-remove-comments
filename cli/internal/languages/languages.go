package languages

import (
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/toml"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

type LangConfig struct {
	Name     string
	Query    string
	Language func() *sitter.Language
}

var byExtension = map[string]LangConfig{
	".js": {
		Name:     "javascript",
		Query:    "(comment) @comment",
		Language: javascript.GetLanguage,
	},
	".mjs": {
		Name:     "javascript",
		Query:    "(comment) @comment",
		Language: javascript.GetLanguage,
	},
	".cjs": {
		Name:     "javascript",
		Query:    "(comment) @comment",
		Language: javascript.GetLanguage,
	},
	".jsx": {
		Name:     "javascript",
		Query:    "(comment) @comment",
		Language: javascript.GetLanguage,
	},
	".ts": {
		Name:     "typescript",
		Query:    "(comment) @comment",
		Language: typescript.GetLanguage,
	},
	".tsx": {
		Name:     "tsx",
		Query:    "(comment) @comment",
		Language: tsx.GetLanguage,
	},
	".lua": {
		Name:     "lua",
		Query:    "(comment) @comment",
		Language: lua.GetLanguage,
	},
	".py": {
		Name:     "python",
		Query:    "(comment) @comment",
		Language: python.GetLanguage,
	},
	".go": {
		Name:     "go",
		Query:    "(comment) @comment",
		Language: golang.GetLanguage,
	},
	".java": {
		Name:     "java",
		Query:    "(line_comment) @comment (block_comment) @comment",
		Language: java.GetLanguage,
	},
	".c": {
		Name:     "c",
		Query:    "(comment) @comment",
		Language: c.GetLanguage,
	},
	".h": {
		Name:     "c",
		Query:    "(comment) @comment",
		Language: c.GetLanguage,
	},
	".cpp": {
		Name:     "cpp",
		Query:    "(comment) @comment",
		Language: cpp.GetLanguage,
	},
	".cc": {
		Name:     "cpp",
		Query:    "(comment) @comment",
		Language: cpp.GetLanguage,
	},
	".cxx": {
		Name:     "cpp",
		Query:    "(comment) @comment",
		Language: cpp.GetLanguage,
	},
	".hpp": {
		Name:     "cpp",
		Query:    "(comment) @comment",
		Language: cpp.GetLanguage,
	},
	".rs": {
		Name:     "rust",
		Query:    "(line_comment) @comment",
		Language: rust.GetLanguage,
	},
	".html": {
		Name:     "html",
		Query:    "(comment) @comment",
		Language: html.GetLanguage,
	},
	".htm": {
		Name:     "html",
		Query:    "(comment) @comment",
		Language: html.GetLanguage,
	},
	".css": {
		Name:     "css",
		Query:    "(comment) @comment",
		Language: css.GetLanguage,
	},
	".yaml": {
		Name:     "yaml",
		Query:    "(comment) @comment",
		Language: yaml.GetLanguage,
	},
	".yml": {
		Name:     "yaml",
		Query:    "(comment) @comment",
		Language: yaml.GetLanguage,
	},
	".toml": {
		Name:     "toml",
		Query:    "(comment) @comment",
		Language: toml.GetLanguage,
	},
	".sh": {
		Name:     "bash",
		Query:    "(comment) @comment",
		Language: bash.GetLanguage,
	},
	".bash": {
		Name:     "bash",
		Query:    "(comment) @comment",
		Language: bash.GetLanguage,
	},
}

func Get(ext string) (LangConfig, bool) {
	cfg, ok := byExtension[ext]
	return cfg, ok
}

func Supported() []string {
	exts := make([]string, 0, len(byExtension))
	for ext := range byExtension {
		exts = append(exts, ext)
	}
	return exts
}
